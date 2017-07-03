# Build Service Adapters for the Cloud Foundry On-Demand Service Broker (ODB) in Golang

---

This is an SDK for writing service adapters for [ODB](https://docs.pivotal.io/svc-sdk/odb) in Golang. It encapsulates the command line invocation handling, parameter parsing, response serialization and error handling so the adapter authors can focus on the service-specific logic in the adapter. This will speed up the time to meeting the service author deliverables outlined [here](https://docs.pivotal.io/svc-sdk/odb/creating.html).

Before creating a service adapter you will need to have [BOSH release](https://bosh.io/docs) for the service that you wish to deploy.

After creating the service adapter and service BOSH release, you will be able to [configure the ODB](https://docs.pivotal.io/svc-sdk/odb/operating.html) provision new dedicated service instances from Cloud Foundry!

---

## Usage 

Please use the SDK tag that matches the ODB release you are targeting.

For example if using ODB 0.15.1 release, use the [0.15.1 SDK tag](https://github.com/pivotal-cf/on-demand-services-sdk/tree/v0.15.1).

### Getting Started

Follow [this guide](https://docs.pivotal.io/svc-sdk/odb/getting-started.html) to try out an example product.

### Examples Service Adapters

Kafka Service Adapter: https://github.com/pivotal-cf-experimental/kafka-example-service-adapter

Redis Service Adapter: https://github.com/pivotal-cf-experimental/redis-example-service-adapter

### Packaging 

To integrate with the ODB we recommend that you package the service adapter in a BOSH release.

#### Examples

Kafka Service Adapter Release: https://github.com/pivotal-cf-experimental/kafka-example-service-adapter-release

Redis Service Adapter Release: https://github.com/pivotal-cf-experimental/redis-example-service-adapter-release

---

### Documentation

SDK Documentation: https://docs.pivotal.io/on-demand-service-broker/creating.html#sdk

On-Demand Services Documentation: https://docs.pivotal.io/on-demand-service-broker/index.html

---

On Demand Services SDK

Copyright (c) 2016 - Present Pivotal Software, Inc. All Rights Reserved. 

This product is licensed to you under the Apache License, Version 2.0 (the "License").  
You may not use this product except in compliance with the License.  

This product may include a number of subcomponents with separate copyright notices 
and license terms. Your use of these subcomponents is subject to the terms and 
conditions of the subcomponent's license, as noted in the LICENSE file.
