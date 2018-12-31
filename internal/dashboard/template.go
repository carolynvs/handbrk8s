package dashboard

const dashboardTemplate = `<html>
<body>
<ul>
{{range .Jobs}}
<li>{{.Name}} ({{ .Duration }}) - {{ .StatusDescription }}</li>
{{end}}
</ul>
</body>
</html>`
