package maven

import (
	"fmt"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/readers"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic/parser"
	"github.com/cantara/buri/version/release"
)

func newestVersion(f filter.Filter, packT string, versions []string) (newest release.Version) {
	for _, v := range versions {
		if packT == "go" {
			if !strings.HasPrefix(v, "v") {
				continue
			}
			//v = strings.TrimPrefix(v, "v")
		}
		if !f.Matches(v) {
			continue
		}

		rv, err := release.Parse(v)
		if err != nil {
			log.WithError(err).Debug("while parsing version")
			continue
		}
		if !newest.IsStrictlySemanticNewer(f, rv) {
			continue
		}
		log.Debug("read newer version")
		newest = rv
	}
	return
}

func Version[T readers.Version[T]](f filter.Filter, url, packageType string) (newestMaven readers.Program[T]) {
	log.Info("maven version", "url", url)
	relVers := newestVersion(f, packageType, GetTableValues(url))
	log.Trace("finding artifact in maven", "version", relVers)
	path := fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), relVers)
	vers := relVers.String()
	if f.Type != release.Type {
		vers = strings.TrimSuffix(vers, "-"+strings.ToUpper(string(f.Type)))
	}
	log.Trace("read", "version", vers, "path", path)
	prog := readers.Program[release.Version]{
		Path:    path,
		Version: relVers,
	}

	if prog.Path == "" {
		log.Fatal("no version found in maven")
	}
	if f.Type == release.Type {
		newestMaven = any(prog).(readers.Program[T]) //This seems bad
		return
	}

	urlParts := strings.Split(url, "/")
	artifactId := urlParts[len(urlParts)-1]
	filenames := GetTableValues(prog.DownloadPath())
	log.Trace("finding maven full versions", "filenames", filenames)
	for _, filename := range filenames {
		log.Trace("non release", "filename", filename)
		version := strings.TrimPrefix(filename, artifactId+"-")
		if packageType == "go" && !strings.HasPrefix(version, "v") {
			continue
		}
		switch packageType {
		case "jar":
			version = strings.TrimSuffix(version, ".jar")
		case "tar":
			version = strings.TrimSuffix(version, ".tgz")
		case "zip":
			version = strings.TrimSuffix(version, ".zip")
		}
		versionString := strings.TrimSuffix(version, "/")
		log.Trace("testing non release version", "maven", versionString)
		versAny, err := parser.Parse(f, versionString)
		if err != nil {
			log.WithError(err).Debug("while parsing version")
			continue
		}
		vers := versAny.(T)
		if !vers.Matches(f) {
			continue
		}
		if err != nil {
			log.WithError(err).Debug("while parsing version")
			continue
		}
		log.Trace("read", "version", vers)
		if newestMaven.Version.IsStrictlySemanticNewer(f, vers) {
			newestMaven = readers.Program[T]{
				Path:    prog.DownloadPath(),
				Version: vers,
			}
			continue
		}

	}
	if newestMaven.Path == "" {
		log.Fatal("no version found in maven")
	}
	log.Trace("new version found in maven", "version", newestMaven.Version, "path", newestMaven.Path)

	return
}
