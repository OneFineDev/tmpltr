package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/OneFineDev/tmpltr/internal/storage"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSourceClient struct {
	mock.Mock
}

func (m *mockSourceClient) CloneSource(ctx context.Context, cloneOpts storage.CloneOpts) (billy.Filesystem, error) {
	args := m.Called(ctx, cloneOpts)
	return args.Get(0).(billy.Filesystem), args.Error(1)
}

func (m *mockSourceClient) GetCurrentSource() *types.Source {
	args := m.Called()
	return args.Get(0).(*types.Source)
}

func (m *mockSourceClient) SetCurrentSource(s *types.Source) {
	m.Called(s)
}

func TestCloneSources(t *testing.T) {
	// Arrange
	tests := []struct {
		name           string
		targetSources  map[string]types.Source
		sourceClients  map[string]services.SourceClient
		expectedErrors []error
	}{
		{
			name: "successful clone",
			targetSources: map[string]types.Source{
				"source1": {
					Alias:      "source1",
					SourceType: types.GitSourceType,
				},
			},
			sourceClients: map[string]services.SourceClient{
				string(types.GitSourceType): func() services.SourceClient {
					client := new(mockSourceClient)
					client.On("SetCurrentSource", mock.Anything).Return()
					client.On("CloneSource", mock.Anything, mock.Anything).Return(memfs.New(), nil)
					return client
				}(),
			},
			expectedErrors: nil,
		},
		// {
		// 	name: "clone failure",
		// 	targetSources: map[string]types.Source{
		// 		"source1": {
		// 			Alias:      "source1",
		// 			SourceType: types.GitSourceType,
		// 		},
		// 	},
		// 	sourceClients: map[string]services.SourceClient{
		// 		string(types.GitSourceType): func() services.SourceClient {
		// 			client := new(mockSourceClient)
		// 			client.On("SetCurrentSource", mock.Anything).Return()
		// 			client.On("CloneSource", mock.Anything, mock.Anything).
		// 				Return(memfs.New(), errors.New("clone error"))
		// 			return client
		// 		}(),
		// 	},
		// 	expectedErrors: []error{
		// 		&package_errors.SourceError{
		// 			Message: "inmem clone failed for {Alias:source1 SourceType:GitSourceType SourceAuthAlias: SourceAuth:<nil>}",
		// 			Err:     errors.New("clone error"),
		// 		},
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			memFs := afero.NewMemMapFs()
			ss := &services.SourceService{
				Logger:        slog.New(slog.DiscardHandler),
				TargetSources: tt.targetSources,
				SourceClients: tt.sourceClients,
				SourcesCommandConfig: &services.SourcesCommandConfig{
					OutputPath: "/output",
				},
			}

			// Act
			resultFs, errs := ss.CloneSources(context.Background(), memFs)

			// Assert
			if tt.expectedErrors == nil {
				assert.NotNil(t, resultFs)
				assert.Empty(t, errs)
			} else {
				assert.Nil(t, resultFs)
				assert.Len(t, errs, len(tt.expectedErrors))
				for i, expectedErr := range tt.expectedErrors {
					assert.EqualError(t, errs[i], expectedErr.Error())
				}
			}
		})
	}
}
