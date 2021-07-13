package ovirtclient

func (o *oVirtClient) GetTemplate(id string) (Template, error) {
	response, err := o.conn.SystemService().TemplatesService().TemplateService(id).Get().Send()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to fetch template %s", id)
	}
	sdkTemplate, ok := response.Template()
	if !ok {
		return nil, newError(ENotFound, "API response contained no template")
	}
	template, err := convertSDKTemplate(sdkTemplate)
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to convert template object")
	}
	return template, nil
}
