package controllers

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog"
)

func TestNewHandler(t *testing.T) {
	logger := zerolog.Logger{}
	c := clamav.Client{}
	type args struct {
		logger *zerolog.Logger
		clamav clamav.Clamaver
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil, nil},
			want: &Handler{nil, nil},
		},
		{
			name: "non nil args",
			args: args{&logger, &c},
			want: &Handler{&c, &logger},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.logger, tt.args.clamav); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockClamav struct{}

var _ clamav.Clamaver = (*MockClamav)(nil)

func (m *MockClamav) Ping(ctx context.Context) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		return []byte("PONG"), nil
	default:
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) Version(ctx context.Context) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		return []byte("ClamAV 1.0.1/26961/Thu Jul  6 07:29:38 2023"), nil
	default:
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) Reload(ctx context.Context) error {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		return nil
	default:
		return dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) Stats(ctx context.Context) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		resp := `POOLS: 1

STATE: VALID PRIMARY
THREADS: live 1  idle 0 max 10 idle-timeout 30
QUEUE: 0 items
	STATS 0.000086 

MEMSTATS: heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 1306.837M pools_total 1306.882M
END`
		return []byte(resp), nil
	case ScenarioStatsErrMarshall:
		resp := `POOLS: POOLS: POOLS: some invalid stats`
		return []byte(resp), nil
	default:
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) VersionCommands(ctx context.Context) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		return []byte("ClamAV 1.0.1/26963/Sat Jul  8 07:27:53 2023| COMMANDS: SCAN QUIT RELOAD PING CONTSCAN VERSIONCOMMANDS VERSION END SHUTDOWN MULTISCAN FILDES STATS IDSESSION INSTREAM DETSTATSCLEAR DETSTATS ALLMATCHSCAN"), nil
	case ScenarioVersionCommandsErrMarshall:
		return []byte("Some unparsable VERSIONCOMMANS output"), nil
	default:
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) Shutdown(ctx context.Context) error {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		return nil
	default:
		return dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) InStream(ctx context.Context, _ io.Reader, _ int64) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	switch scenario {
	case ScenarioNoError:
		return []byte("stream: OK"), nil
	case ScenarioErrVirusFound:
		return []byte("stream: Win.Test.EICAR_HDB-1 FOUND"), clamav.ErrVirusFound
	default:
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func dispatchErrFromScenario(scenario MockScenario) error {
	switch scenario {
	case ScenarioNetError:
		return &net.OpError{Err: errors.New("network error")}
	case ScenarioErrUnknownCommand:
		return clamav.ErrUnknownCommand
	case ScenarioErrUnknownResponse:
		return clamav.ErrUnknownResponse
	case ScenarioErrUnexpectedResponse:
		return clamav.ErrUnexpectedResponse
	case ScenarioErrScanFileSizeLimitExceeded:
		return clamav.ErrScanFileSizeLimitExceeded
	case ScenarioErrVirusFound:
		return clamav.ErrVirusFound
	default:
		return nil
	}
}

type MockScenario string

var (
	ScenarioNoError                      MockScenario = "noerror"
	ScenarioNetError                     MockScenario = "neterror"
	ScenarioErrUnknownCommand            MockScenario = "unknowncommand"
	ScenarioErrUnknownResponse           MockScenario = "unknownresponse"
	ScenarioErrUnexpectedResponse        MockScenario = "unexpectedresponse"
	ScenarioErrScanFileSizeLimitExceeded MockScenario = "scanfilesizelimitexceeded"

	ScenarioStatsErrMarshall           MockScenario = "statserrmarshall"
	ScenarioVersionCommandsErrMarshall MockScenario = "versioncommandserrmarshall"

	ScenarioErrVirusFound MockScenario = "virusfound"
)
