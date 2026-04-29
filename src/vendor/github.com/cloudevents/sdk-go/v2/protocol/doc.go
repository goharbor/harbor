/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

/*
Package protocol defines interfaces to decouple the client package
from protocol implementations.

Most event sender and receiver applications should not use this
package, they should use the client package. This package is for
infrastructure developers implementing new transports, or intermediary
components like importers, channels or brokers.

Available protocols:

* HTTP (using net/http)
* Kafka (using github.com/Shopify/sarama)
* AMQP (using pack.ag/amqp)
* Go Channels
* Nats
* Nats Streaming (stan)
* Google PubSub
*/
package protocol
