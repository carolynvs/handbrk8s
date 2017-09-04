package api

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
)

// ProcessTemplate substitutes the supplied values into a yaml template.
func ProcessTemplate(yamlTemplate string, values interface{}) (yaml []byte, err error) {
	t, err := template.New("").Parse(yamlTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse the yaml template\n%s", yamlTemplate)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, values)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to execute the yaml template with the values:\n%#v", values)
	}

	return buf.Bytes(), nil
}

// DeserializeObject reads a k8s object from yaml.
func DeserializeObject(yaml []byte) (runtime.Object, error) {
	serializer := api.Codecs.UniversalDeserializer()

	obj, _, err := serializer.Decode(yaml, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot decode the yaml into a k8s runtime object\n%s", string(yaml))
	}
	return obj, err
}

// SerializeObject returns the yaml representation of a k8s object.
func SerializeObject(obj runtime.Object) (string, error) {
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	var buf bytes.Buffer
	err := serializer.Encode(obj, &buf)
	if err != nil {
		return "", errors.Wrapf(err, "unable to serialize object:\n%#v", obj)
	}
	return buf.String(), nil
}
