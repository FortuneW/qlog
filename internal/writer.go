package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"sync/atomic"
)

type (
	Writer interface {
		Close() error
		Trace(v any)
		Debug(v any)
		Warn(v any)
		Error(v any)
		Info(v any)
		AccessRecord(v any)
		WriteRawString(v string)
	}

	atomicWriter struct {
		writer Writer
		lock   sync.RWMutex
	}

	comboWriter struct {
		writers []Writer
	}

	concreteWriter struct {
		serverLog  io.WriteCloser
		managerLog io.WriteCloser
	}

	emptyWriter struct{}
)

// NewWriter creates a new Writer with the given io.Writer.
func NewWriter(w io.Writer) Writer {
	lw := newLogWriter(log.New(w, "", flags))

	return &concreteWriter{
		serverLog:  lw,
		managerLog: lw,
	}
}

func (w *atomicWriter) Load() Writer {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.writer
}

func (w *atomicWriter) Store(v Writer) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.writer = v
}

func (w *atomicWriter) StoreIfNil(v Writer) Writer {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.writer == nil {
		w.writer = v
	}

	return w.writer
}

func (w *atomicWriter) Swap(v Writer) Writer {
	w.lock.Lock()
	defer w.lock.Unlock()
	old := w.writer
	w.writer = v
	return old
}

func (c comboWriter) Close() error {
	var be BatchError
	for _, w := range c.writers {
		be.Add(w.Close())
	}
	return be.Err()
}

func (c comboWriter) Trace(v any) {
	for _, w := range c.writers {
		w.Trace(v)
	}
}
func (c comboWriter) Debug(v any) {
	for _, w := range c.writers {
		w.Debug(v)
	}
}

func (c comboWriter) Error(v any) {
	for _, w := range c.writers {
		w.Error(v)
	}
}

func (c comboWriter) Warn(v any) {
	for _, w := range c.writers {
		w.Warn(v)
	}
}

func (c comboWriter) Info(v any) {
	for _, w := range c.writers {
		w.Info(v)
	}
}

func (c comboWriter) AccessRecord(v any) {
	for _, w := range c.writers {
		w.AccessRecord(v)
	}
}
func (c comboWriter) WriteRawString(v string) {
	for _, w := range c.writers {
		w.WriteRawString(v)
	}
}

func newConsoleWriter() Writer {
	outLog := newLogWriter(log.New(os.Stdout, "", flags))
	return &concreteWriter{
		serverLog:  outLog,
		managerLog: outLog,
	}
}

func newFileWriter(c LogConf) (Writer, error) {
	var err error
	var opts []LogOption
	var serverLog io.WriteCloser
	var managerLog io.WriteCloser

	if len(c.ServerLogDir) == 0 || len(c.ManagerLogDir) == 0 {
		return nil, ErrLogPathNotSet
	}

	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}
	if c.MaxBackups > 0 {
		opts = append(opts, WithMaxBackups(c.MaxBackups))
	}
	if c.MaxSize > 0 {
		opts = append(opts, WithMaxSize(c.MaxSize))
	}

	opts = append(opts, WithRotation(c.Rotation))

	managerFile := path.Join(c.ManagerLogDir, c.ServiceName+"_"+managerFilename)
	serverFile := path.Join(c.ServerLogDir, c.ServiceName+"_"+serverFilename)

	handleOptions(opts)
	setupLogLevel(c)

	if serverLog, err = createOutput(serverFile); err != nil {
		return nil, err
	}

	if managerLog, err = createOutput(managerFile); err != nil {
		return nil, err
	}

	return &concreteWriter{
		serverLog:  serverLog,
		managerLog: managerLog,
	}, nil
}

func (w *concreteWriter) Close() error {

	if err := w.serverLog.Close(); err != nil {
		return err
	}

	if err := w.managerLog.Close(); err != nil {
		return err
	}

	return nil
}

func (w *concreteWriter) Trace(v any) {
	output(w.serverLog, LevelTrace, v)
}

func (w *concreteWriter) Debug(v any) {
	output(w.serverLog, LevelDebug, v)
}

func (w *concreteWriter) Error(v any) {
	output(w.serverLog, LevelError, v)
}

func (w *concreteWriter) Warn(v any) {
	output(w.serverLog, LevelWarn, v)
}

func (w *concreteWriter) Info(v any) {
	output(w.serverLog, LevelInfo, v)
}

func (w *concreteWriter) AccessRecord(v any) {
	output(w.managerLog, levelAccessRecord, v)
}

func (w *concreteWriter) WriteRawString(v string) {
	writer := w.serverLog
	if writer == nil {
		log.Print(v)
		return
	}

	if _, err := writer.Write([]byte(v)); err != nil {
		log.Println(err.Error())
	}
}

func output(writer io.Writer, level string, val any) {
	// only truncate string content, don't know how to truncate the values of other types.
	if v, ok := val.(string); ok {
		maxLen := atomic.LoadUint32(&maxContentLength)
		if maxLen > 0 && len(v) > int(maxLen) {
			val = v[:maxLen]
		}
	}
	writePlainAny(writer, level, val)
}

func GetOutputStringFormatted(level string, val any) string {
	switch v := val.(type) {
	case string:
		text := formatPlainText(level, v)
		return text.String()
	case error:
		text := formatPlainText(level, v.Error())
		return text.String()
	case fmt.Stringer:
		text := formatPlainText(level, v.String())
		return text.String()
	default:
		text := formatPlainValue(level, v)
		return text.String()
	}
}

func writePlainAny(writer io.Writer, level string, val any) {
	switch v := val.(type) {
	case string:
		writePlainText(writer, level, v)
	case error:
		writePlainText(writer, level, v.Error())
	case fmt.Stringer:
		writePlainText(writer, level, v.String())
	default:
		writePlainValue(writer, level, v)
	}
}

func formatPlainText(level, msg string) bytes.Buffer {
	var buf bytes.Buffer
	if level != "" && level != levelAccessRecord {
		buf.WriteByte('[')
		buf.WriteString(level)
		buf.WriteByte(']')
		buf.WriteByte(plainEncodingSep)
	}
	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(msg)
	buf.WriteByte('\n')
	return buf
}

func writePlainText(writer io.Writer, level, msg string) {
	buf := formatPlainText(level, msg)

	if writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := writer.Write(buf.Bytes()); err != nil {
		log.Println(err.Error())
	}
}

func formatPlainValue(level string, val any) bytes.Buffer {
	var buf bytes.Buffer
	if level != "" && level != levelAccessRecord {
		buf.WriteByte('[')
		buf.WriteString(level)
		buf.WriteByte(']')
		buf.WriteByte(plainEncodingSep)
	}

	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)

	// 兜底用json表示对象的字符串
	_ = json.NewEncoder(&buf).Encode(val)

	buf.WriteByte('\n')
	return buf
}

func writePlainValue(writer io.Writer, level string, val any) {
	buf := formatPlainValue(level, val)

	if writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := writer.Write(buf.Bytes()); err != nil {
		log.Println(err.Error())
	}
}

func newEmptyWriter() Writer {
	return &emptyWriter{}
}

func (w *emptyWriter) Close() error            { return nil }
func (w *emptyWriter) Trace(v any)             {}
func (w *emptyWriter) Debug(v any)             {}
func (w *emptyWriter) Warn(v any)              {}
func (w *emptyWriter) Error(v any)             {}
func (w *emptyWriter) Info(v any)              {}
func (w *emptyWriter) AccessRecord(v any)      {}
func (w *emptyWriter) WriteRawString(v string) {}
