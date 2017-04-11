package notifiers

import (
	"context"
	"testing"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"
)

func init() {
	registCreator("test", testNotifierCreator)
}

type testNotifier struct {
}

var testNotifierCreator = func(model models.Notifier) (Notifier, error) {
	return &testNotifier{}, nil
}

func (tn *testNotifier) Send(ctx context.Context, report models.Report) error {
	return nil
}

func TestDefaultManager(t *testing.T) {
	n, err := DefaultManager.CreateNotifier(models.Notifier{
		Type: "test",
	})
	assert.NoError(t, err)
	assert.NoError(t, n.Send(context.Background(), models.Report{}))
}
