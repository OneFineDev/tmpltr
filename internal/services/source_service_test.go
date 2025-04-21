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
	ctx := context.Background()
	memFs := afero.NewMemMapFs()
	logger := slog.New(slog.DiscardHandler)

	mockSourceClient := new(mockSourceClient)
	mockSource := types.Source{
		Alias: "test-source",
	}
	mockSourceClient.On("CloneSource", mock.Anything, mock.Anything).Return(memfs.New(), nil)

	sourceService := &services.SourceService{
		SourcesCommandConfig: &services.SourcesCommandConfig{
			OutputPath: "/output",
		},
		Logger:        logger,
		TargetSources: map[string]types.Source{"test-source": mockSource},
		SourceClients: map[string]services.SourceClient{"test-source": mockSourceClient},
	}

	// Act
	billyChan, errChan := sourceService.CloneSources(ctx, memFs)

	// Assert
	var billyFs billy.Filesystem
	var err error
	select {
	case billyFs = <-billyChan:
		assert.NotNil(t, billyFs, "Expected a non-nil billy.Filesystem")
	case err = <-errChan:
		assert.Fail(t, "Unexpected error received", err)
	}

	// Ensure channels are closed
	_, billyChanOpen := <-billyChan
	_, errChanOpen := <-errChan
	assert.False(t, billyChanOpen, "Expected billyChan to be closed")
	assert.False(t, errChanOpen, "Expected errChan to be closed")
}
