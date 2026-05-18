package spectra

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func writeZipArchive(outPath, dataPath string, manifest *Manifest, shards []shardPayload) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := zip.NewWriter(out)
	defer writer.Close()

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := writeZipEntry(writer, manifestFile, manifestBytes); err != nil {
		return err
	}

	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}
	if err := writeZipEntry(writer, dataFile, dataBytes); err != nil {
		return err
	}

	for _, shard := range shards {
		path := filepath.Join(jsonShardDir, shard.tableName+".json")
		if err := writeZipEntry(writer, path, shard.jsonBytes); err != nil {
			return err
		}
	}
	return nil
}

func writeZipEntry(writer *zip.Writer, name string, body []byte) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, bytes.NewReader(body))
	return err
}
