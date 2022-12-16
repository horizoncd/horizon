/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templaterepo

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

const (
	sep = string(filepath.Separator)

	// ChartfileName is the default Chart file name.
	ChartfileName = "Chart.yaml"
	// ValuesfileName is the default values file name.
	ValuesfileName = "values.yaml"
	// SchemafileName is the default values schema file name.
	SchemafileName = "values.schema.json"
	// TemplatesDir is the relative directory name for templates.
	TemplatesDir = "templates"
	// ChartsDir is the relative directory name for charts dependencies.
	ChartsDir = "charts"
	// TemplatesTestsDir is the relative directory name for tests.
	TemplatesTestsDir = TemplatesDir + sep + "tests"
	// IgnorefileName is the name of the Helm ignore file.
	IgnorefileName = ".helmignore"
	// IngressFileName is the name of the example ingress file.
	IngressFileName = TemplatesDir + sep + "ingress.yaml"
	// DeploymentName is the name of the example deployment file.
	DeploymentName = TemplatesDir + sep + "deployment.yaml"
	// ServiceName is the name of the example service file.
	ServiceName = TemplatesDir + sep + "service.yaml"
	// ServiceAccountName is the name of the example serviceaccount file.
	ServiceAccountName = TemplatesDir + sep + "serviceaccount.yaml"
	// HorizontalPodAutoscalerName is the name of the example hpa file.
	HorizontalPodAutoscalerName = TemplatesDir + sep + "hpa.yaml"
	// NotesName is the name of the example NOTES.txt file.
	NotesName = TemplatesDir + sep + "NOTES.txt"
	// HelpersName is the name of the example helpers file.
	HelpersName = TemplatesDir + sep + "_helpers.tpl"
	// TestConnectionName is the name of the example test file.
	TestConnectionName = TemplatesTestsDir + sep + "test-connection.yaml"
)

func WriteTarContents(out *tar.Writer, c *chart.Chart, prefix string) error {
	base := filepath.Join(prefix, c.Name())

	// Pull out the dependencies of a v1 Chart, since there's no way
	// to tell the serializer to skip a field for just this use case
	savedDependencies := c.Metadata.Dependencies
	if c.Metadata.APIVersion == chart.APIVersionV1 {
		c.Metadata.Dependencies = nil
	}
	// Save Chart.yaml
	cdata, err := yaml.Marshal(c.Metadata)
	if c.Metadata.APIVersion == chart.APIVersionV1 {
		c.Metadata.Dependencies = savedDependencies
	}
	if err != nil {
		return err
	}
	if err := writeToTar(out, filepath.Join(base, ChartfileName), cdata); err != nil {
		return err
	}

	// Save Chart.lock
	// TODO: remove the APIVersion check when APIVersionV1 is not used anymore
	if c.Metadata.APIVersion == chart.APIVersionV2 {
		if c.Lock != nil {
			ldata, err := yaml.Marshal(c.Lock)
			if err != nil {
				return err
			}
			if err := writeToTar(out, filepath.Join(base, "Chart.lock"), ldata); err != nil {
				return err
			}
		}
	}

	// Save values.yaml
	for _, f := range c.Raw {
		if f.Name == ValuesfileName {
			if err := writeToTar(out, filepath.Join(base, ValuesfileName), f.Data); err != nil {
				return err
			}
		}
	}

	// Save values.schema.json if it exists
	if c.Schema != nil {
		if !json.Valid(c.Schema) {
			return errors.New("Invalid JSON in " + SchemafileName)
		}
		if err := writeToTar(out, filepath.Join(base, SchemafileName), c.Schema); err != nil {
			return err
		}
	}

	// Save templates
	for _, f := range c.Templates {
		n := filepath.Join(base, f.Name)
		if err := writeToTar(out, n, f.Data); err != nil {
			return err
		}
	}

	// Save files
	for _, f := range c.Files {
		n := filepath.Join(base, f.Name)
		if err := writeToTar(out, n, f.Data); err != nil {
			return err
		}
	}

	// Save dependencies
	for _, dep := range c.Dependencies() {
		if err := WriteTarContents(out, dep, filepath.Join(base, ChartsDir)); err != nil {
			return err
		}
	}
	return nil
}

// writeToTar writes a single file to a tar archive.
func writeToTar(out *tar.Writer, name string, body []byte) error {
	// TODO: Do we need to create dummy parent directory names if none exist?
	h := &tar.Header{
		Name:    filepath.ToSlash(name),
		Mode:    0644,
		Size:    int64(len(body)),
		ModTime: time.Now(),
	}
	if err := out.WriteHeader(h); err != nil {
		return err
	}
	_, err := out.Write(body)
	return err
}

func ChartSerialize(c *chart.Chart, out io.Writer) error {
	gzipper := gzip.NewWriter(out)
	gzipper.Header.Comment = "Helm"

	twriter := tar.NewWriter(gzipper)
	defer func() {
		_ = twriter.Close()
		_ = gzipper.Close()
	}()
	err := WriteTarContents(twriter, c, "")
	if err != nil {
		return perror.Wrap(herrors.ErrWriteFailed,
			fmt.Sprintf("writing chart to buffer failed: %s", err.Error()))
	}
	return nil
}
