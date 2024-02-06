# PinGoTrace

PinGoTrace has been created to help network/systems engineers query or monitor the availability of another node on the network in a more efficient way than the standard Command Prompt tool.

## DOMAIN/IP PARSER
Parses hostnames or IPv4s from the text.

## DNS/PTR
For each hostname or IPv4 parsed, performs DNS or PTR lookup and displays results.

## DNS/PTR to IP
For each DNS or PTR resolution, displays only the corresponding IPv4 address.

## Infinity PING
Parses the input and issues continuous Ping for each DNS or PTR resolution. Supports up to 30 continuous Ping runs concurrently.

## TRACE
Parses the input and issues Traceroute for the first DNS or PTR resolution.

## PINGOTRACE
Parses the input and issues Traceroute for the first DNS or PTR resolution. Upon completion, starts continuous Ping against each live hop.

## Infinity TRACE
Parses the input and issues continuous Traceroute for the first DNS or PTR resolution. A 3-second delay is between each Traceroute.

## IPCONFIG
Displays IP information of the workstation.

## CLEAR
Deletes previously entered text from the display.
