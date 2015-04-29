function FindProxyForURL(url, host) {
	if (shExpMatch(host, "*.example.com"))
	{
		return "DIRECT";
	}
	if (isInNet(host, "10.0.0.0", "255.255.248.0"))
	{
		return "PROXY fastproxy.example.com:8080";
	}
	return "PROXY proxy.example.com:8080; DIRECT";
}
