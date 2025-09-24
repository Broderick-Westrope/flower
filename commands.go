package main

type WorkCmd struct {
	Task string `arg:"" help:"Task description"`
}

func (w *WorkCmd) Run() error {
	return handleWork(w.Task)
}

type BreakCmd struct{}

func (b *BreakCmd) Run() error {
	return handleBreak()
}

type ResumeCmd struct{}

func (r *ResumeCmd) Run() error {
	return handleResume()
}

type StopCmd struct{}

func (s *StopCmd) Run() error {
	return handleStop()
}

type StatusCmd struct{}

func (s *StatusCmd) Run() error {
	return handleStatus()
}

type LogCmd struct{}

func (l *LogCmd) Run() error {
	return handleLog()
}
