// this is a new logger interface for mattermost

package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// ensures that values can be recorded on a Context object, and that the data in question is serialized as a part of the log message
func TestSerializeContext(t *testing.T) {
	t.Run("Context values test", func(t *testing.T) {
		ctx := context.Background()

		expectedUserID := "some-fake-user-id"
		ctx = WithUserID(ctx, expectedUserID)

		expectedRequestID := "some-fake-request-id"
		ctx = WithRequestID(ctx, expectedRequestID)

		serialized := serializeContext(ctx)

		userID, ok := serialized["user-id"]
		if !ok {
			t.Error("UserID was not serialized")
		}
		if userID != expectedUserID {
			t.Errorf("UserID = %v, want %v", userID, expectedUserID)
		}

		requestID, ok := serialized["request-id"]
		if !ok {
			t.Error("RequestID was not serialized")
		}
		if requestID != expectedRequestID {
			t.Errorf("RequestID = %v, want %v", requestID, expectedRequestID)
		}
	})
}

// ensures that an entire log message with an empty context can be properly serialized into a JSON object
func TestSerializeLogMessageEmptyContext(t *testing.T) {
	emptyContext := context.Background()

	var logMessage = "This is a log message"
	var serialized = serializeLogMessage(emptyContext, logMessage)

	type LogMessage struct {
		Context map[string]string
		Logger  string
		Message string
	}
	var deserialized LogMessage
	json.Unmarshal([]byte(serialized), &deserialized)

	if len(deserialized.Context) != 0 {
		t.Error("Context is non-empty")
	}
	var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
	if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
		t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
	}
	if deserialized.Message != logMessage {
		t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, logMessage)
	}
}

// ensures that an entire log message with a populated context can be properly serialized into a JSON object
func TestSerializeLogMessagePopulatedContext(t *testing.T) {
	populatedContext := context.Background()

	populatedContext = WithRequestID(populatedContext, "foo")
	populatedContext = WithUserID(populatedContext, "bar")

	var logMessage = "This is a log message"
	var serialized = serializeLogMessage(populatedContext, logMessage)

	type LogMessage struct {
		Context map[string]string
		Logger  string
		Message string
	}
	var deserialized LogMessage
	json.Unmarshal([]byte(serialized), &deserialized)

	if len(deserialized.Context) != 2 {
		t.Error("Context is non-empty")
	}
	if deserialized.Context["request-id"] != "foo" {
		t.Errorf("Invalid request-id %v. Expected %v", deserialized.Context["request-id"], "foo")
	}
	if deserialized.Context["user-id"] != "bar" {
		t.Errorf("Invalid user-id %v. Expected %v", deserialized.Context["user-id"], "bar")
	}
	var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
	if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
		t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
	}
	if deserialized.Message != logMessage {
		t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, logMessage)
	}
}

// ensures that a debug message is passed through to the underlying logger as expected
func TestDebugc(t *testing.T) {
	t.Run("Debugc test", func(t *testing.T) {
		// inject a "mocked" debug method that captures the first argument that is passed to it
		var capture string
		oldDebug := debug
		defer func() { debug = oldDebug }()
		type WrapperType func() string
		debug = func(format interface{}, args ...interface{}) {
			// the code that we're testing passes a closure to the debug method, so we have to execute it to get the actual message back
			if f, ok := format.(func() string); ok {
				capture = WrapperType(f)()
			} else {
				t.Error("First parameter passed to Debug is not a closure")
			}
		}

		// log something
		emptyContext := context.Background()
		var logMessage = "Some log message"
		Debugc(emptyContext, logMessage)

		// check to see that the message is logged to the underlying log system, in this case our mock method
		type LogMessage struct {
			Context map[string]string
			Logger  string
			Message string
		}
		var deserialized LogMessage
		json.Unmarshal([]byte(capture), &deserialized)

		if len(deserialized.Context) != 0 {
			t.Error("Context is non-empty")
		}
		var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
		if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
			t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
		}
		if deserialized.Message != logMessage {
			t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, logMessage)
		}
	})
}

// ensures that a debug message is passed through to the underlying logger as expected
func TestDebugf(t *testing.T) {
	t.Run("Debugf test", func(t *testing.T) {
		// inject a "mocked" debug method that captures the first argument that is passed to it
		var capture string
		oldDebug := debug
		defer func() { debug = oldDebug }()
		type WrapperType func() string
		debug = func(format interface{}, args ...interface{}) {
			// the code that we're testing passes a closure to the debug method, so we have to execute it to get the actual message back
			if f, ok := format.(func() string); ok {
				capture = WrapperType(f)()
			} else {
				t.Error("First parameter passed to Debug is not a closure")
			}
		}

		// log something
		formatString := "Some %v message"
		param := "log"
		Debugf(formatString, param)

		// check to see that the message is logged to the underlying log system, in this case our mock method
		type LogMessage struct {
			Context map[string]string
			Logger  string
			Message string
		}
		var deserialized LogMessage
		json.Unmarshal([]byte(capture), &deserialized)

		if len(deserialized.Context) != 0 {
			t.Error("Context is non-empty")
		}
		var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
		if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
			t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
		}

		expected := fmt.Sprintf(formatString, param)
		if deserialized.Message != expected {
			t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, expected)
		}
	})
}

// ensures that an info message is passed through to the underlying logger as expected
func TestInfoc(t *testing.T) {
	t.Run("Infoc test", func(t *testing.T) {
		// inject a "mocked" info method that captures the first argument that is passed to it
		var capture string
		oldInfo := info
		defer func() { info = oldInfo }()
		type WrapperType func() string
		info = func(format interface{}, args ...interface{}) {
			// the code that we're testing passes a closure to the info method, so we have to execute it to get the actual message back
			if f, ok := format.(func() string); ok {
				capture = WrapperType(f)()
			} else {
				t.Error("First parameter passed to Info is not a closure")
			}
		}

		// log something
		emptyContext := context.Background()
		var logMessage = "Some log message"
		Infoc(emptyContext, logMessage)

		// check to see that the message is logged to the underlying log system, in this case our mock method
		type LogMessage struct {
			Context map[string]string
			Logger  string
			Message string
		}
		var deserialized LogMessage
		json.Unmarshal([]byte(capture), &deserialized)

		if len(deserialized.Context) != 0 {
			t.Error("Context is non-empty")
		}
		var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
		if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
			t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
		}
		if deserialized.Message != logMessage {
			t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, logMessage)
		}
	})
}

// ensures that an info message is passed through to the underlying logger as expected
func TestInfof(t *testing.T) {
	t.Run("Infof test", func(t *testing.T) {
		// inject a "mocked" info method that captures the first argument that is passed to it
		var capture string
		oldInfo := info
		defer func() { info = oldInfo }()
		type WrapperType func() string
		info = func(format interface{}, args ...interface{}) {
			// the code that we're testing passes a closure to the info method, so we have to execute it to get the actual message back
			if f, ok := format.(func() string); ok {
				capture = WrapperType(f)()
			} else {
				t.Error("First parameter passed to Info is not a closure")
			}
		}

		// log something
		format := "Some %v message"
		param := "log"
		Infof(format, param)

		// check to see that the message is logged to the underlying log system, in this case our mock method
		type LogMessage struct {
			Context map[string]string
			Logger  string
			Message string
		}
		var deserialized LogMessage
		json.Unmarshal([]byte(capture), &deserialized)

		if len(deserialized.Context) != 0 {
			t.Error("Context is non-empty")
		}
		var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
		if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
			t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
		}

		expected := fmt.Sprintf(format, param)
		if deserialized.Message != expected {
			t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, expected)
		}
	})
}

// ensures that an error message is passed through to the underlying logger as expected
func TestErrorc(t *testing.T) {
	t.Run("Errorc test", func(t *testing.T) {
		// inject a "mocked" error method that captures the first argument that is passed to it
		var capture string
		oldError := err
		defer func() { err = oldError }()
		type WrapperType func() string
		err = func(format interface{}, args ...interface{}) error {
			// the code that we're testing passes a closure to the error method, so we have to execute it to get the actual message back
			if f, ok := format.(func() string); ok {
				capture = WrapperType(f)()
			} else {
				t.Error("First parameter passed to Error is not a closure")
			}

			// the code under test doesn't care about this return value
			return errors.New(capture)
		}

		// log something
		emptyContext := context.Background()
		var logMessage = "Some log message"
		Errorc(emptyContext, logMessage)

		// check to see that the message is logged to the underlying log system, in this case our mock method
		type LogMessage struct {
			Context map[string]string
			Logger  string
			Message string
		}
		var deserialized LogMessage
		json.Unmarshal([]byte(capture), &deserialized)

		if len(deserialized.Context) != 0 {
			t.Error("Context is non-empty")
		}
		var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
		if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
			t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
		}
		if deserialized.Message != logMessage {
			t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, logMessage)
		}
	})
}

// ensures that an error message is passed through to the underlying logger as expected
func TestErrorf(t *testing.T) {
	t.Run("Errorf test", func(t *testing.T) {
		// inject a "mocked" error method that captures the first argument that is passed to it
		var capture string
		oldError := err
		defer func() { err = oldError }()
		type WrapperType func() string
		err = func(format interface{}, args ...interface{}) error {
			// the code that we're testing passes a closure to the error method, so we have to execute it to get the actual message back
			if f, ok := format.(func() string); ok {
				capture = WrapperType(f)()
			} else {
				t.Error("First parameter passed to Error is not a closure")
			}

			// the code under test doesn't care about this return value
			return errors.New(capture)
		}

		// log something
		format := "Some %v message"
		param := "log"
		Errorf(format, param)

		// check to see that the message is logged to the underlying log system, in this case our mock method
		type LogMessage struct {
			Context map[string]string
			Logger  string
			Message string
		}
		var deserialized LogMessage
		json.Unmarshal([]byte(capture), &deserialized)

		if len(deserialized.Context) != 0 {
			t.Error("Context is non-empty")
		}
		var expectedLoggerSuffix = "/platform/utils/logger/logger_test.go"
		if !strings.HasSuffix(deserialized.Logger, expectedLoggerSuffix) {
			t.Errorf("Invalid logger %v. Expected logger to have suffix %v", deserialized.Logger, expectedLoggerSuffix)
		}

		expected := fmt.Sprintf(format, param)
		if deserialized.Message != expected {
			t.Errorf("Invalid log message %v. Expected %v", deserialized.Message, expected)
		}
	})
}
