// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package mtail_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/testutil"
)

func TestTruncatedLogRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, test := range mtail.LogWatcherTestTable {
		t.Run(fmt.Sprintf("%s %v", test.PollInterval, test.EnableFsNotify), func(t *testing.T) {
			tmpDir, rmTmpDir := testutil.TestTempDir(t)
			defer rmTmpDir()

			logDir := path.Join(tmpDir, "logs")
			progDir := path.Join(tmpDir, "progs")
			testutil.FatalIfErr(t, os.Mkdir(logDir, 0700))
			testutil.FatalIfErr(t, os.Mkdir(progDir, 0700))

			m, stopM := mtail.TestStartServer(t, test.PollInterval, test.EnableFsNotify, mtail.ProgramPath(progDir), mtail.LogPathPatterns(logDir+"/log"))
			defer stopM()

			logCountCheck := m.ExpectMetricDeltaWithDeadline("log_count", 1)

			logFile := path.Join(logDir, "log")
			f := testutil.TestOpenFile(t, logFile)
			if !test.EnableFsNotify {
				m.PollWatched()
			}

			{
				linesCountCheck := m.ExpectMetricDeltaWithDeadline("lines_total", 1)
				testutil.WriteString(t, f, "1\n")
				if !test.EnableFsNotify {
					m.PollWatched()
				}
				linesCountCheck()
			}
			err := f.Close()
			testutil.FatalIfErr(t, err)
			f, err = os.OpenFile(logFile, os.O_TRUNC|os.O_RDWR, 0600)
			testutil.FatalIfErr(t, err)
			// Ensure the server notices the truncate
			if !test.EnableFsNotify {
				m.PollWatched()
			}
			{
				linesCountCheck := m.ExpectMetricDeltaWithDeadline("lines_total", 1)
				testutil.WriteString(t, f, "2\n")
				if !test.EnableFsNotify {
					m.PollWatched()
				}
				linesCountCheck()
			}
			logCountCheck()
		})
	}
}
