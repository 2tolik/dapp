package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/flant/yaml.v2"

	"github.com/flant/dapp/pkg/util"
)

func ParseDimgs(dappfilePath string) ([]*Dimg, error) {
	dappfileRenderContent, err := parseDappfileYaml(dappfilePath)
	if err != nil {
		return nil, err
	}

	dappfileRenderPath, err := dumpDappfileRender(dappfilePath, dappfileRenderContent)
	if err != nil {
		return nil, err
	}

	docs, err := splitByDocs(dappfileRenderContent, dappfileRenderPath)
	if err != nil {
		return nil, err
	}

	dimgs, err := splitByDimgs(docs, dappfileRenderContent, dappfileRenderPath)
	if err != nil {
		return nil, err
	}

	return dimgs, nil
}

func dumpDappfileRender(dappfilePath string, dappfileRenderContent string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dappfileNameParts := strings.Split(path.Base(dappfilePath), ".")
	dappfileRenderNameParts := []string{}
	dappfileRenderNameParts = append(dappfileRenderNameParts, dappfileNameParts[0:len(dappfileNameParts)-1]...)
	dappfileRenderNameParts = append(dappfileRenderNameParts, "render", dappfileNameParts[len(dappfileNameParts)-1])
	dappfileRenderPath := path.Join(wd, fmt.Sprintf(".%s", strings.Join(dappfileRenderNameParts, ".")))

	dappfileRenderFile, err := os.OpenFile(dappfileRenderPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	dappfileRenderFile.Write([]byte(dappfileRenderContent))
	dappfileRenderFile.Close()

	return dappfileRenderPath, nil
}

func splitByDocs(dappfileRenderContent string, dappfileRenderPath string) ([]*Doc, error) {
	var docs []*Doc
	var line int
	for _, docContent := range splitContent([]byte(dappfileRenderContent)) {
		if !emptyDocContent(docContent) {
			docs = append(docs, &Doc{
				Line:           line,
				Content:        docContent,
				RenderFilePath: dappfileRenderPath,
			})
		}

		contentLines := bytes.Split(docContent, []byte("\n"))
		if string(contentLines[len(contentLines)-1]) == "" {
			contentLines = contentLines[0 : len(contentLines)-1]
		}
		line += len(contentLines) + 1
	}

	return docs, nil
}

// TODO: переделать на ParseFiles вместо Parse
func parseDappfileYaml(dappfilePath string) (string, error) {
	data, err := ioutil.ReadFile(dappfilePath)
	if err != nil {
		return "", err
	}

	tmpl := template.New("dappfile")
	tmpl.Funcs(funcMap(tmpl))
	if _, err := tmpl.Parse(string(data)); err != nil {
		return "", err
	}
	return executeTemplate(tmpl, "dappfile", map[string]interface{}{"Files": Files{filepath.Dir(dappfilePath)}})
}

func funcMap(tmpl *template.Template) template.FuncMap {
	funcMap := sprig.TxtFuncMap()
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		return executeTemplate(tmpl, name, data)
	}
	return funcMap
}

func executeTemplate(tmpl *template.Template, name string, data interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type Files struct {
	HomePath string
}

func (f Files) Get(path string) string {
	b, err := ioutil.ReadFile(filepath.Join(f.HomePath, path))
	if err != nil {
		return ""
	}
	return string(b)
}

func splitContent(content []byte) (docsContents [][]byte) {
	const (
		stateLineBegin   = "stateLineBegin"
		stateRegularLine = "stateRegularLine"
		stateDocDash1    = "stateDocDash1"
		stateDocDash2    = "stateDocDash2"
		stateDocDash3    = "stateDocDash3"
		stateDocSpaces   = "stateDocSpaces"
		stateDocComment  = "stateDocComment"
	)

	state := stateLineBegin
	var docStartIndex, separatorLength int
	var docContent []byte
	var index int
	var ch byte
	for index, ch = range content {
		switch ch {
		case '-':
			switch state {
			case stateLineBegin:
				separatorLength = 1
				state = stateDocDash1
			case stateDocDash1, stateDocDash2:
				separatorLength += 1

				switch state {
				case stateDocDash1:
					state = stateDocDash2
				case stateDocDash2:
					state = stateDocDash3
				}
			default:
				state = stateRegularLine
			}
		case '\n':
			switch state {
			case stateDocDash3, stateDocSpaces, stateDocComment:
				if docStartIndex == index-separatorLength {
					docContent = []byte{}
				} else {
					docContent = content[docStartIndex : index-separatorLength]
				}
				docsContents = append(docsContents, docContent)
				docStartIndex = index + 1
			}
			separatorLength = 0
			state = stateLineBegin
		case ' ', '\r', '\t':
			switch state {
			case stateDocDash3, stateDocSpaces:
				separatorLength += 1
				state = stateDocSpaces
			case stateDocComment:
				separatorLength += 1
			default:
				state = stateRegularLine
			}
		case '#':
			switch state {
			case stateDocDash3, stateDocSpaces, stateDocComment:
				separatorLength += 1
				state = stateDocComment
			default:
				state = stateRegularLine
			}
		default:
			switch state {
			case stateDocComment:
				separatorLength += 1
			default:
				state = stateRegularLine
			}
		}
	}

	if docStartIndex != index+1 {
		switch state {
		case stateDocDash3, stateDocSpaces, stateDocComment:
			if docStartIndex == index-separatorLength {
				docContent = []byte{}
			} else {
				docContent = content[docStartIndex : index-separatorLength]
			}
		default:
			docContent = content[docStartIndex:]
		}
		docsContents = append(docsContents, docContent)
	}

	return docsContents
}

func emptyDocContent(content []byte) bool {
	const (
		stateRegular = 0
		stateComment = 1
	)

	state := stateRegular
	for _, ch := range content {
		switch ch {
		case '#':
			state = stateComment
		case '\n':
			state = stateRegular
		case ' ', '\r', '\t':
		default:
			if state == stateRegular {
				return false
			}
		}
	}
	return true
}

func splitByDimgs(docs []*Doc, dappfileRenderContent string, dappfileRenderPath string) ([]*Dimg, error) {
	rawDimgs, err := splitByRawDimgs(docs)
	if err != nil {
		return nil, err
	}

	var dimgs []*Dimg
	var artifacts []*DimgArtifact

	for _, rawDimg := range rawDimgs {
		if rawDimg.Type() == "dimgs" {
			if sameDimgs, err := rawDimg.ToDimgDirectives(); err != nil {
				return nil, err
			} else {
				dimgs = append(dimgs, sameDimgs...)
			}
		} else {
			if dimgArtifact, err := rawDimg.ToDimgArtifactDirective(); err != nil {
				return nil, err
			} else {
				artifacts = append(artifacts, dimgArtifact)
			}
		}
	}

	if len(dimgs) == 0 {
		return nil, NewConfigError(fmt.Sprintf("No dimgs defined, at least one dimg required!\n\n%s:\n\n```\n%s```\n", dappfileRenderPath, dappfileRenderContent))
	}

	if err = validateDimgsNames(dimgs); err != nil {
		return nil, err
	}

	if err = validateArtifactsNames(artifacts); err != nil {
		return nil, err
	}

	if err = associateArtifacts(dimgs, artifacts); err != nil {
		return nil, err
	}

	if err = associateFromDimg(dimgs); err != nil {
		return nil, err
	}

	return dimgs, nil
}

func validateDimgsNames(dimgs []*Dimg) error {
	dimgNames := map[string]*Dimg{}
	for _, dimg := range dimgs {
		if d, ok := dimgNames[dimg.Name]; ok {
			return NewConfigError(fmt.Sprintf("Conflict between dimgs names!\n\n%s%s\n", DumpConfigDoc(d.Raw.Doc), DumpConfigDoc(dimg.Raw.Doc)))
		} else {
			dimgNames[dimg.Name] = dimg
		}
	}
	return nil
}

func validateArtifactsNames(artifacts []*DimgArtifact) error {
	artifactsNames := map[string]*DimgArtifact{}
	for _, artifact := range artifacts {
		if a, ok := artifactsNames[artifact.Name]; ok {
			return NewConfigError(fmt.Sprintf("Conflict between artifacts names!\n\n%s%s\n", DumpConfigDoc(a.Raw.Doc), DumpConfigDoc(artifact.Raw.Doc)))
		} else {
			artifactsNames[artifact.Name] = artifact
		}
	}
	return nil
}

func associateArtifacts(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	var artifactImports []*ArtifactImport

	for _, dimg := range dimgs {
		for _, relatedDimgInterface := range dimg.RelatedDimgs() {
			switch relatedDimgInterface.(type) {
			case *Dimg:
				artifactImports = append(artifactImports, relatedDimgInterface.(*Dimg).Import...)
			case *DimgArtifact:
				artifactImports = append(artifactImports, relatedDimgInterface.(*DimgArtifact).Import...)
			}
		}
	}

	for _, artifactDimg := range artifacts {
		for _, relatedDimgInterface := range artifactDimg.RelatedDimgs() {
			switch relatedDimgInterface.(type) {
			case *Dimg:
				artifactImports = append(artifactImports, relatedDimgInterface.(*Dimg).Import...)
			case *DimgArtifact:
				artifactImports = append(artifactImports, relatedDimgInterface.(*DimgArtifact).Import...)
			}
		}
	}

	for _, artifactImport := range artifactImports {
		if err := artifactImport.AssociateArtifact(artifacts); err != nil {
			return err
		}
	}

	return nil
}

func associateFromDimg(dimgs []*Dimg) error {
	for _, dimg := range dimgs {
		if err := dimg.AssociateFromDimg(dimgs); err != nil {
			return err
		}
	}

	return nil
}

func splitByRawDimgs(docs []*Doc) ([]*RawDimg, error) {
	var rawDimgs []*RawDimg
	ParentStack = util.NewStack()
	for _, doc := range docs {
		dimg := &RawDimg{Doc: doc}
		err := yaml.Unmarshal(doc.Content, &dimg)
		if err != nil {
			return nil, newYamlUnmarshalError(err, doc)
		}
		rawDimgs = append(rawDimgs, dimg)
	}

	return rawDimgs, nil
}

func newYamlUnmarshalError(err error, doc *Doc) error {
	switch err.(type) {
	case *ConfigError:
		return err
	default:
		message := err.Error()
		reg, err := regexp.Compile("line ([0-9]+)")
		if err != nil {
			return err
		}

		res := reg.FindStringSubmatch(message)

		if len(res) == 2 {
			line, err := strconv.Atoi(res[1])
			if err != nil {
				return err
			}

			message = reg.ReplaceAllString(message, fmt.Sprintf("line %d", line+doc.Line))
		}
		return NewDetailedConfigError(message, nil, doc)
	}
}
