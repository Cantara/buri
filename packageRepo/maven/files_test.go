package maven

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	filenames := parse(nexus2_buri_html)
	if len(filenames) != numVersionsInTestData {
		t.Errorf("Missmatch amount of filenames(%d) returned and versions parsed(%d)\n%s", len(filenames), numVersionsInTestData, filenames)
	}
}

func TestUniquify(t *testing.T) {
	in := []string{"1", "2", "1"}
	expect := []string{"1", "2"}
	uniqueified := uniqueify(in)
	if !reflect.DeepEqual(expect, uniqueified) {
		t.Errorf("expected %v got %v", expect, uniqueified)
	}
}

const numVersionsInTestData = 52
const nexus2_buri_html = `<html><head>
    <title>Index of /repositories/releases/no/cantara/gotools/buri</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">

    <link rel="icon" type="image/png" href="http://nexus.cantara.no/favicon.png">
    <!--[if IE]>
    <link rel="SHORTCUT ICON" href="http://nexus.cantara.no/favicon.ico"/>
    <![endif]-->

    <link rel="stylesheet" href="http://nexus.cantara.no/static/css/Sonatype-content.css?2.12.1-01" type="text/css" media="screen" title="no title" charset="utf-8">
  </head>
  <body>
    <h1>Index of /repositories/releases/no/cantara/gotools/buri</h1>
    <table cellspacing="10">
      <tbody><tr>
        <th align="left">Name</th>
        <th>Last Modified</th>
        <th>Size</th>
        <th>Description</th>
      </tr>
      <tr>
        <td><a href="../">Parent Directory</a></td>
      </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.10.1/">v0.10.1/</a></td>
            <td>Mon Apr 24 06:09:36 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.10.2/">v0.10.2/</a></td>
            <td>Mon Apr 24 06:58:33 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.10.3/">v0.10.3/</a></td>
            <td>Mon Apr 24 09:43:38 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.10.4/">v0.10.4/</a></td>
            <td>Mon Apr 24 10:08:30 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.10.5/">v0.10.5/</a></td>
            <td>Wed Apr 26 03:47:38 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.10.6/">v0.10.6/</a></td>
            <td>Wed Apr 26 08:58:48 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.0/">v0.11.0/</a></td>
            <td>Fri Apr 28 08:59:39 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.1/">v0.11.1/</a></td>
            <td>Fri Apr 28 09:16:35 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.11/">v0.11.11/</a></td>
            <td>Tue Jul 11 15:26:56 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.12/">v0.11.12/</a></td>
            <td>Tue Jul 11 18:54:33 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.13/">v0.11.13/</a></td>
            <td>Tue Jul 11 19:16:33 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.14/">v0.11.14/</a></td>
            <td>Tue Jul 11 19:17:35 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.15/">v0.11.15/</a></td>
            <td>Tue Jul 11 19:23:37 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.2/">v0.11.2/</a></td>
            <td>Fri Apr 28 10:05:37 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.3/">v0.11.3/</a></td>
            <td>Fri Apr 28 11:23:39 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.4/">v0.11.4/</a></td>
            <td>Fri Apr 28 11:44:33 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.5/">v0.11.5/</a></td>
            <td>Fri Apr 28 13:15:35 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.6/">v0.11.6/</a></td>
            <td>Fri Apr 28 13:22:40 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.7/">v0.11.7/</a></td>
            <td>Fri Apr 28 13:29:39 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.8/">v0.11.8/</a></td>
            <td>Fri Apr 28 13:51:40 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.11.9/">v0.11.9/</a></td>
            <td>Fri Apr 28 14:26:34 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.2.1/">v0.2.1/</a></td>
            <td>Tue Jan 18 10:12:24 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.0/">v0.3.0/</a></td>
            <td>Tue Jan 18 10:14:23 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.1/">v0.3.1/</a></td>
            <td>Tue Jan 18 12:21:46 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.2/">v0.3.2/</a></td>
            <td>Tue Jan 18 14:40:55 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.3/">v0.3.3/</a></td>
            <td>Tue Jan 18 16:08:26 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.5/">v0.3.5/</a></td>
            <td>Tue Jan 18 16:52:12 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.4.0/">v0.4.0/</a></td>
            <td>Thu Jan 20 13:37:22 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.4.1/">v0.4.1/</a></td>
            <td>Thu Jan 20 17:46:19 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.4.2/">v0.4.2/</a></td>
            <td>Thu Jan 20 17:50:53 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.4.3/">v0.4.3/</a></td>
            <td>Thu Jan 20 19:50:24 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.5.0/">v0.5.0/</a></td>
            <td>Sat Feb 05 20:38:20 UTC 2022</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.6.0/">v0.6.0/</a></td>
            <td>Sat Feb 18 10:33:21 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.7.0/">v0.7.0/</a></td>
            <td>Sun Feb 19 02:59:19 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.0/">v0.8.0/</a></td>
            <td>Thu Mar 16 06:45:22 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.1/">v0.8.1/</a></td>
            <td>Tue Mar 21 08:23:29 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.2/">v0.8.2/</a></td>
            <td>Tue Mar 21 09:58:33 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.3/">v0.8.3/</a></td>
            <td>Tue Mar 21 15:36:52 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.4/">v0.8.4/</a></td>
            <td>Wed Mar 22 06:56:28 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.5/">v0.8.5/</a></td>
            <td>Wed Mar 22 14:11:28 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.6/">v0.8.6/</a></td>
            <td>Wed Mar 22 14:18:56 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.7/">v0.8.7/</a></td>
            <td>Wed Mar 22 14:31:26 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.8.8/">v0.8.8/</a></td>
            <td>Sat Mar 25 04:04:29 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.0/">v0.9.0/</a></td>
            <td>Sat Mar 25 11:04:31 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.1/">v0.9.1/</a></td>
            <td>Sat Mar 25 17:40:30 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.2/">v0.9.2/</a></td>
            <td>Sun Mar 26 10:13:29 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.3/">v0.9.3/</a></td>
            <td>Sun Mar 26 18:56:26 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.4/">v0.9.4/</a></td>
            <td>Sun Mar 26 19:00:26 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.5/">v0.9.5/</a></td>
            <td>Sun Mar 26 19:03:26 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.6/">v0.9.6/</a></td>
            <td>Tue Mar 28 12:26:34 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.7/">v0.9.7/</a></td>
            <td>Thu Apr 13 12:24:37 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
                  <tr>
            <td><a href="http://nexus.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.9.8/">v0.9.8/</a></td>
            <td>Sun Apr 16 09:17:26 UTC 2023</td>
            <td align="right">
                              &nbsp;
                          </td>
            <td></td>
          </tr>
            </tbody></table>


</body></html>`
