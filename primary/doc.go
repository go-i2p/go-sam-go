// Package primary provides PRIMARY session management for sharing I2P tunnels across multiple subsessions.
//
// PRIMARY sessions allow multiple subsessions (stream, datagram, datagram2, datagram3, raw)
// to share a single set of I2P tunnels, reducing resource usage and tunnel setup overhead.
// Each subsession operates independently while using the master session's tunnels.
//
// Key features:
//   - Single tunnel setup for multiple subsessions
//   - Mixed subsession types (stream, datagram, raw)
//   - Independent subsession lifecycle management
//   - Reduced resource usage and setup time
//   - SAMv3.3 PRIMARY protocol compliance
//
// Primary session creation requires 2-5 minutes for I2P tunnel establishment. Subsessions
// attach quickly since tunnels are already established. Use generous timeouts for initial
// PRIMARY session creation.
//
// Basic usage:
//
//	sam, err := common.NewSAM("127.0.0.1:7656")
//	primary, err := primary.NewPrimarySession(sam, "master", keys, []string{"inbound.length=1"})
//	defer primary.Close()
//	streamSub, err := primary.NewStreamSubsession("stream-1")
//	datagramSub, err := primary.NewDatagramSubsession("dgram-1")
//
// See also: Package stream, datagram, datagram2, datagram3, raw for individual session types.
package primary
