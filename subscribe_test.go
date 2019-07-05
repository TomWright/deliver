package deliver

import (
	"errors"
	"fmt"
	"testing"
)

func TestSubscribeOptions_Validate(t *testing.T) {
	tests := []struct{
		desc string
		options SubscribeOptions
		err error
		postCheck func(options SubscribeOptions) error
	}{
		{
			desc: "valid with missing group",
			options: SubscribeOptions{
				Group: "",
				ConsumeFn: func(messageType string, messageBytes []byte) error {
					return nil
				},
				Types: []string{"test"},
				IgnoreErrors: true,
			},
			postCheck: func(options SubscribeOptions) error {
				if exp, got := "default", options.Group; exp != got {
					return fmt.Errorf("expected group `%s`, got `%s`", exp, got)
				}
				return nil
			},

		},
		{
			desc: "missing consumer fn",
			options: SubscribeOptions{
				Group: "",
				Types: []string{"test"},
				IgnoreErrors: true,
			},
			err: errors.New("missing consumer function"),
		},
		{
			desc: "missing types",
			options: SubscribeOptions{
				Group: "",
				ConsumeFn: func(messageType string, messageBytes []byte) error {
					return nil
				},
				IgnoreErrors: true,
			},
			err: errors.New("no message types to subscribe to"),
		},
		{
			desc: "wrongly given err channel",
			options: SubscribeOptions{
				Group: "",
				ConsumeFn: func(messageType string, messageBytes []byte) error {
					return nil
				},
				Types: []string{"test"},
				IgnoreErrors: true,
				Errors: make(chan error),
			},
			err: errors.New("ignore errors is on but error channel was provided"),
		},
		{
			desc: "missing err channel",
			options: SubscribeOptions{
				Group: "",
				ConsumeFn: func(messageType string, messageBytes []byte) error {
					return nil
				},
				Types: []string{"test"},
				IgnoreErrors: false,
			},
			err: errors.New("ignore errors is off but no error channel was provided"),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			err := tc.options.Validate()

			if tc.err == nil && err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
			if tc.err != nil && err == nil {
				t.Errorf("expected error `%s` but got none", tc.err)
				return
			}
			if tc.err != nil && err != nil && tc.err.Error() != err.Error() {
				t.Errorf("expected error `%s` but got `%s`", tc.err, err)
				return
			}

			if tc.postCheck != nil {
				if err := tc.postCheck(tc.options); err != nil {
					t.Errorf("post check failed: %s", err)
					return
				}
			}
		})
	}
}