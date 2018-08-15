package backends

import (
	"github.com/sirupsen/logrus"
	)

//var log logging.Logger
var log *logrus.Logger

func UseLogger (logger *logrus.Logger) {//(logger logging.Logger) {
	log = logger
}
