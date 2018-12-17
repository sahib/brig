// Package events is the event notification subsystem of brig.
// It uses the backend's capabilities (in case of IPFS we use pubsub)
// to publish and subscribe to a topic of events. If an event was received
// it is forwarded to the caller side in order to react on it.
package events
