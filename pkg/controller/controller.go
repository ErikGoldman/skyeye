// package controller implements high-level logic for Ground-Controlled Interception (GCI)
package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/lithammer/shortuuid/v3"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

var (
	lowestAltitude  = unit.Length(0)
	highestAltitude = unit.Length(100000) * unit.Foot
)

type Call struct {
	Context context.Context
	Call    any
}

func NewCall(ctx context.Context, call any) Call {
	return Call{
		Context: ctx,
		Call:    call,
	}
}

// Controller handles requests for GCI service.
type Controller interface {
	// Run starts the controller's control loops. It should be called exactly once. It blocks until the context is canceled.
	// The controller publishes responses to the given channel.
	Run(ctx context.Context, out chan<- Call)
	// HandleAlphaCheck handles an ALPHA CHECK by reporting the position of the requesting aircraft.
	HandleAlphaCheck(context.Context, *brevity.AlphaCheckRequest)
	// HandleBogeyDope handles a BOGEY DOPE by reporting the closest enemy group to the requesting aircraft.
	HandleBogeyDope(context.Context, *brevity.BogeyDopeRequest)
	// HandleCheckIn handles an ambiguous CHECK IN by asking the player to clarify their call.
	HandleCheckIn(context.Context, *brevity.CheckInRequest)
	// HandleDeclare handles a DECLARE by reporting information about the target group.
	HandleDeclare(context.Context, *brevity.DeclareRequest)
	// HandlePicture handles a PICTURE by reporting a tactical air picture.
	HandlePicture(context.Context, *brevity.PictureRequest)
	// HandleRadioCheck handles a RADIO CHECK by responding to the requesting aircraft.
	HandleRadioCheck(context.Context, *brevity.RadioCheckRequest)
	// HandleShopping handles a SHOPPING request... by not implementing it, since it's not an air-to-air call!
	HandleShopping(context.Context, *brevity.ShoppingRequest)
	// HandleSnaplock handles a SNAPLOCK by reporting information about the target group.
	HandleSnaplock(context.Context, *brevity.SnaplockRequest)
	// HandleSpiked handles a SPIKED by reporting any enemy groups in the direction of the radar spike.
	HandleSpiked(context.Context, *brevity.SpikedRequest)
	// HandleTripwire handles a TRIPWIRE... by not implementing it LOL
	HandleTripwire(context.Context, *brevity.TripwireRequest)
	// HandleUnableToUnderstand handles requests where the wake word was recognized but the request could not be understood, by asking players on the channel to repeat their message.
	HandleUnableToUnderstand(context.Context, *brevity.UnableToUnderstandRequest)
}

type controller struct {
	// coalition this controller serves.
	coalition coalitions.Coalition

	// scope provides information about the airspace.
	scope radar.Radar

	// srsClient is used to to check if relevant friendly aircraft are on frequency before broadcasting calls.
	srsClient simpleradio.Client

	// enableAutomaticPicture enables automatic picture broadcasts.
	enableAutomaticPicture bool
	// pictureBroadcastInterval is the interval at which the controller broadcasts a tactical air picture.
	pictureBroadcastInterval time.Duration
	// pictureBroadcastDeadline is the time at which the controller will broadcast the next tactical air picture.
	pictureBroadcastDeadline time.Time
	// wasLastPictureClean tracks if the most recently broadcast picture was clean, so that the controller can avoid
	// repeatedly broadcasting clean pictures.
	wasLastPictureClean bool

	// enableThreatMonitoring enables automatic threat calls.
	enableThreatMonitoring bool
	// threatCooldowns tracks the next time a threat call should be published for each threat.
	threatCooldowns *cooldownTracker
	// threatMonitoringCooldown is the interval between threat calls for the same threat.
	threatMonitoringCooldown time.Duration
	// threatMonitoringRequiresSRS enforces that threat calls are only broadcast when the relevant friendly aircraft are on frequency.
	threatMonitoringRequiresSRS bool

	// merges tracks which contacts are in the merge.
	merges *mergeTracker

	// calls is the channel to publish responses and calls to.
	calls chan<- Call
}

func New(
	rdr radar.Radar,
	srsClient simpleradio.Client,
	coalition coalitions.Coalition,
	enableAutomaticPicture bool,
	pictureBroadcastInterval time.Duration,
	enableThreatMonitoring bool,
	threatMonitoringCooldown time.Duration,
	threatMonitoringRequiresSRS bool,
) Controller {
	return &controller{
		coalition:                   coalition,
		scope:                       rdr,
		srsClient:                   srsClient,
		enableAutomaticPicture:      enableAutomaticPicture,
		pictureBroadcastInterval:    pictureBroadcastInterval,
		pictureBroadcastDeadline:    time.Now().Add(pictureBroadcastInterval),
		enableThreatMonitoring:      enableThreatMonitoring,
		threatMonitoringCooldown:    threatMonitoringCooldown,
		threatCooldowns:             newCooldownTracker(threatMonitoringCooldown),
		threatMonitoringRequiresSRS: threatMonitoringRequiresSRS,
		merges:                      newMergeTracker(),
	}
}

// Run implements [Controller.Run].
func (c *controller) Run(ctx context.Context, calls chan<- Call) {
	c.calls = calls

	log.Info().Msg("attaching callbacks")
	c.scope.SetFadedCallback(c.handleFaded)
	c.scope.SetRemovedCallback(c.handleRemoved)

	c.broadcastSunrise(ctx)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("detaching callbacks")
			c.scope.SetFadedCallback(nil)
			c.scope.SetRemovedCallback(nil)
			c.scope.SetStartedCallback(nil)
			return
		case <-ticker.C:
			c.broadcastMerges(traces.WithTraceID(ctx, shortuuid.New()))
			c.broadcastThreats(traces.WithTraceID(ctx, shortuuid.New()))
			if c.enableAutomaticPicture && time.Now().After(c.pictureBroadcastDeadline) {
				c.broadcastPicture(traces.WithTraceID(ctx, shortuuid.New()), &log.Logger, false)
			}
		}
	}
}

func (c *controller) broadcastSunrise(ctx context.Context) {
	frequencies := make([]unit.Frequency, 0)
	for _, rf := range c.srsClient.Frequencies() {
		frequencies = append(frequencies, rf.Frequency)
	}
	c.calls <- NewCall(traces.WithTraceID(ctx, shortuuid.New()), brevity.SunriseCall{Frequencies: frequencies})
}

// findCallsign uses fuzzy matching to find a trackfile for the given callsign.
// Any matching callsign is returned, along with any trackfile and a bool indicating
// if a valid trackfile with a non-zero location was found.
func (c *controller) findCallsign(callsign string) (string, *trackfiles.Trackfile, bool) {
	logger := log.With().Str("parsedCallsign", callsign).Logger()
	foundCallsign, trackfile := c.scope.FindCallsign(callsign, c.coalition)
	if trackfile == nil {
		logger.Info().Msg("no trackfile found for callsign")
		return "", nil, false
	}
	logger = logger.With().Str("foundCallsign", foundCallsign).Logger()
	if trackfile.IsLastKnownPointZero() {
		logger.Info().Msg("found trackfile for callsign but without location")
		return foundCallsign, trackfile, false
	}
	logger.Debug().Msg("found trackfile for callsign")
	return foundCallsign, trackfile, true
}

func (c *controller) remove(id uint64) {
	log.Debug().Uint64("id", id).Msg("removing ID from controller state tracking")
	c.threatCooldowns.remove(id)
	c.merges.remove(id)
}

func (c *controller) reset() {
	c.threatCooldowns.reset()
	c.merges.reset()
}
