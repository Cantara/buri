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
	params := GetParamsURL("<td>(.+)</td>", url)
	//log.Debug(params)
	//var programs []readers.Program[release.Version]
	var newestMavenDir readers.Program[release.Version]
	for i := 1; i+1 < len(params); i++ {
		urlPars := GetParams("<a href=\"(.+)\">(.+)</a>", params[i])
		if len(urlPars) != 2 {
			//log.Fatal("Wrong number of urls in path to version")
			continue
		}
		if !strings.HasSuffix(urlPars[0], "/") {
			continue
		}
		if packageType == "go" && !strings.HasPrefix(urlPars[1], "v") { //Could be removed if you don't want go specific selection
			continue
		}
		//log.Println(urlPars)
		/*
			t, err := time.Parse("Mon Jan 02 15:04:05 MST 2006", params[i+1])
			if err != nil {
				log.Fatal(err)
			}
		*/

		var path string
		if strings.HasPrefix(urlPars[0], "http") {
			path = urlPars[0]
		} else {
			path = fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), strings.TrimPrefix(urlPars[0], "/"))
		}
		versionString := strings.Split(strings.TrimSuffix(urlPars[1], "/"), "-")[0]
		/*
			versAny, err := parser.Parse(f, versionString)
			if err != nil {
				log.WithError(err).Debug("while parsing version")
				continue
			}
			vers := versAny.(T)
			if !vers.Matches(f) {
				continue
			}
		*/
		vers, err := release.Parse(versionString)
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
	if newestMavenDir.Path == "" {
		log.Fatal("no version found in maven")
	}
	if f.Type == release.Type {
		newestMavenDir.Path = strings.ReplaceAll(newestMavenDir.Path, "service/rest/repository/browse/", "repository/")
		newestMaven = any(newestMavenDir).(readers.Program[T])
		return
	}
	params = GetParamsURL("<td>(.+)</td>", strings.TrimSuffix(newestMavenDir.Path, "/"))
	for i := 1; i+1 < len(params); i++ {
		urlPars := GetParams("<a href=\"(.+)\">(.+)</a>", params[i])
		if len(urlPars) != 2 {
			//log.Fatal("Wrong number of urls in path to version")
			continue
		}
		if !strings.HasSuffix(urlPars[0], "/") {
			continue
		}
		if packageType == "go" && !strings.HasPrefix(urlPars[1], "v") { //Could be removed if you don't want go specific selection
			continue
		}
		//log.Println(urlPars)
		/*
			t, err := time.Parse("Mon Jan 02 15:04:05 MST 2006", params[i+1])
			if err != nil {
				log.Fatal(err)
			}
		*/
		/*
			var path string
			if strings.HasPrefix(urlPars[0], "http") {
				path = urlPars[0]
			} else {
				path = fmt.Sprintf("%s/%s", strings.TrimSuffix(strings.ReplaceAll(newestMavenDir.Path, "service/rest/repository/browse/", "repository/"),
					"/"), strings.TrimPrefix(urlPars[0], "/"))
			}
		*/
		versionString := strings.TrimSuffix(urlPars[1], "/")
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
				Path:    strings.ReplaceAll(newestMavenDir.Path, "service/rest/repository/browse/", "repository/"),
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
