package roles

var ArgoCSVTemplate = `# Owner
p, role:{{ .Tenant.Name }}-tenant-owner, repositories, get, *, allow
p, role:{{ .Tenant.Name }}-tenant-owner, applicationsets, *, {{ .Tenant.Name }}/*, allow
p, role:{{ .Tenant.Name }}-tenant-owner, applications, *, {{ .Tenant.Name }}/*, allow
p, role:{{ .Tenant.Name }}-tenant-owner, clusters, list, {{ .Endpoint }}, allow
p, role:{{ .Tenant.Name }}-tenant-owner, clusters, get, {{ .Endpoint }}, allow
p, role:{{ .Tenant.Name }}-tenant-owner, projects, get, {{ .Tenant.Name }}, allow
p, role:{{ .Tenant.Name }}-tenant-owner, logs, get, *, allow
p, role:{{ .Tenant.Name }}-tenant-owner, exec, create, */*, deny

# Maintainer
p, role:{{ .Tenant.Name }}-tenant-maintainer, repositories, get, *, allow
p, role:{{ .Tenant.Name }}-tenant-maintainer, clusters, get, *, deny
p, role:{{ .Tenant.Name }}-tenant-maintainer, clusters, get, {{ .Tenant.Name }}, allow

# Assign Owner
{{- range .Tenant.Spec.Owners }}
  {{- if or (eq .Kind "User") (eq .Kind "Group") }}
g, {{ .Name }}, role:{{ $.Tenant.Name }}-tenant-owner
  {{- end }}
{{- end }}

# Assign Maintainer
{{- range .Tenant.Spec.AdditionalRoleBindings }}
  {{- if eq .ClusterRoleName "tenant:maintainer" }}
	{{- range .Subjects }}
	  {{- if or (eq .Kind "User") (eq .Kind "Group") }}	
g, {{ .Name }}, role:{{ $.Tenant.Name }}-tenant-maintainer
	  {{- end }}
	{{- end }}
  {{- end }}
{{- end }}`
