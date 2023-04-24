package maven

import (
	"fmt"
	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/readers"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic/parser"
	"github.com/cantara/buri/version/release"
	"strings"
)

func Version[T readers.Version[T]](f filter.Filter, url, packageType string) (newestMaven readers.Program[T]) {
	log.Info("maven version", "url", url)
	filenames := GetFileNames(url)
	log.Trace("finding maven base versions", "filenames", filenames)
	var newestMavenDir readers.Program[release.Version]
	for _, filename := range filenames {
		/* Nexus 2 has Nexus 3 does not
		if !strings.HasSuffix(filename, "/") {
			continue
		}
		*/
		filename = strings.TrimSuffix(filename, "/")
		filterMatch := filename
		if packageType == "go" {
			if !strings.HasPrefix(filename, "v") {
				continue
			}
			filterMatch = strings.TrimPrefix(filterMatch, "v")
		}
		if !f.Matches(filterMatch) {
			continue
		}

		path := fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), filename)

		if f.Type != release.Type {
			filename = strings.TrimSuffix(filename, "-"+strings.ToUpper(string(f.Type)))
		}
		vers, err := release.Parse(filename)
		if err != nil {
			log.WithError(err).Debug("while parsing version")
			continue
		}
		log.Trace("read", "version", vers, "path", path)
		if newestMavenDir.Version.IsStrictlySemanticNewer(f, vers) {
			log.Debug("read newer version")
			newestMavenDir = readers.Program[release.Version]{
				Path:    path,
				Version: vers,
			}
			continue
		}
	}
	if newestMavenDir.Path == "" {
		log.Fatal("no version found in maven")
	}
	if f.Type == release.Type {
		newestMavenDir.Path = strings.ReplaceAll(newestMavenDir.Path, "service/rest/repository/browse/", "repository/") + "/"
		newestMaven = any(newestMavenDir).(readers.Program[T])
		return
	}

	urlParts := strings.Split(url, "/")
	artifactId := urlParts[len(urlParts)-1]
	filenames = GetFileNames(newestMavenDir.Path)
	log.Trace("finding maven full versions", "filenames", filenames)
	for _, filename := range filenames {
		log.Trace("non release", "filename", filename)
		version := strings.TrimPrefix(filename, artifactId+"-")
		if packageType == "go" && !strings.HasPrefix(version, "v") {
			continue
		}
		if strings.HasSuffix(packageType, "jar") {
			version = strings.TrimSuffix(version, ".jar")
		}
		if strings.HasPrefix(packageType, "zip") {
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
				Path:    strings.ReplaceAll(newestMavenDir.Path, "service/rest/repository/browse/", "repository/") + "/",
				Version: vers,
				//updatedTime: t,
			}
			continue
		}

		/*
			programs = append(programs, readers.Program[release.Version]{
				Path:    path,
				Version: vers,
				//updatedTime: t,
			})
		*/
	}
	/*
		for i, p := range programs {
			if newestP == nil {
				newestP = &programs[i]
				continue
			}
			log.Debug("testing", "filter", f, "v1", newestP.Version, "v2", p.Version)
			if newestP.Version.IsStrictlySemanticNewer(f, p.Version) {
				log.Debug("was newer")
				newestP = &programs[i]
			}
		}
	*/
	if newestMaven.Path == "" {
		log.Fatal("no version found in maven")
	}
	log.Trace("new version found in maven", "version", newestMaven.Version, "path", newestMaven.Path)

	return
}
