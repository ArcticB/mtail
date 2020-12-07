// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.
// +build integration

package mtail_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/testutil"
)

func TestMultipleLinesInOneWrite(t *testing.T) {
	tmpDir, rmTmpDir := testutil.TestTempDir(t)
	defer rmTmpDir()

	logDir := path.Join(tmpDir, "logs")
	progDir := path.Join(tmpDir, "progs")
	err := os.Mkdir(logDir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(progDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	logFile := path.Join(logDir, "log")

	f := testutil.TestOpenFile(t, logFile)

	m, stopM := mtail.TestStartServer(t, 0, false, mtail.ProgramPath(progDir), mtail.LogPathPatterns(logDir+"/log"))
	defer stopM()

	{
		n, err := f.WriteString("line 1\n")
		if err != nil {
			t.Fatal(err)
		}
		glog.Infof("Wrote %d bytes", n)
	}

	lineCountCheck := mtail.ExpectMetricDeltaWithDeadline(t, m.Addr(), "lines_total", 2, time.Minute)

	{

		n, err := f.WriteString("line 2\nline 3\n")
		if err != nil {
			t.Fatal(err)
		}
		glog.Infof("Wrote %d bytes", n)
	}

	lineCountCheck()
}
