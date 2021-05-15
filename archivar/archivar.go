package archivar

import (
	"github.com/rwese/archivar/archivar/archiver"
	"github.com/rwese/archivar/archivar/filter"
	filterMiddleware "github.com/rwese/archivar/archivar/filter/middleware"
	"github.com/rwese/archivar/archivar/gatherer"
	"github.com/rwese/archivar/archivar/job"
	"github.com/rwese/archivar/archivar/processor"
	processorMiddleware "github.com/rwese/archivar/archivar/processor/middleware"
	"github.com/sirupsen/logrus"
)

type Archivar struct {
	logger *logrus.Logger
	jobs   []job.Job
	config Config
}

// ConfigSub is used with an empty Interface to give to the respective package for reflection
type ConfigSub struct {
	Interval int
	Type     string
	Config   interface{}
}

type Config struct {
	Settings struct {
		DefaultInterval int
		Log             struct {
			Debugging bool
		}
	}
	Archivers  map[string]ConfigSub
	Gatherers  map[string]ConfigSub
	Filters    map[string]ConfigSub
	Processors map[string]ConfigSub
	Jobs       map[string]job.JobConfig
}

func New(config Config, logger *logrus.Logger) Archivar {
	s := Archivar{
		logger: logger,
		config: config,
	}

	for jobName, job := range config.Jobs {
		s.addJob(jobName, job)
	}

	return s
}

func (s *Archivar) addJob(jobName string, j job.JobConfig) {
	interval := s.config.Settings.DefaultInterval
	if j.Interval != 0 {
		interval = j.Interval
	}

	c := s.config.Archivers[j.Archiver]
	archiver := archiver.New(c.Type, c.Config, s.logger)

	for _, processorName := range j.Processors {
		c = s.config.Processors[processorName]
		p := processor.New(c.Type, c.Config, s.logger)
		archiver = processorMiddleware.New(archiver, p)
	}

	for _, filterName := range j.Filters {
		c = s.config.Filters[filterName]
		f := filter.New(c.Type, c.Config, s.logger)
		archiver = filterMiddleware.New(archiver, f)
	}

	c = s.config.Gatherers[j.Gatherer]
	gatherer := gatherer.New(c.Type, c.Config, archiver, s.logger)
	s.jobs = append(s.jobs, job.Job{
		Name:     jobName,
		Interval: interval,
		Gatherer: gatherer,
	})
}
