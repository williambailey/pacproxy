console.warn("Default pac configuration loaded")
console.warn("Unless another configuration is loaded only DIRECT connections will be specified");

function FindProxyForURL(url, host)
{
  return "DIRECT";
}
