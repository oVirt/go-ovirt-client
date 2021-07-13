package ovirtclient

func (o *oVirtClient) ListTemplates() ([]Template, error) {
	response, err := o.conn.SystemService().TemplatesService().List().Send()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to list templates")
	}
	sdkTemplates, ok := response.Templates()
	if !ok {
		return nil, newError(ENotFound, "host list response didn't contain hosts")
	}
	result := make([]Template, len(sdkTemplates.Slice()))
	for i, sdkTemplate := range sdkTemplates.Slice() {
		result[i], err = convertSDKTemplate(sdkTemplate)
		if err != nil {
			return nil, wrap(err, EBug, "failed to convert host %d in listing", i)
		}
	}
	return result, nil
}
