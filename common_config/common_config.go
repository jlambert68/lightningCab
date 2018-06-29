package common_config

// gRPC-ports
const GrpcTollServer_address = "127.0.0.1"
const GrpcTollServer_port = ":50651"

const GrpcTollHardwareServer_address = "127.0.0.1"
const GrpcTollHardwareServer_port = ":50650"

const GrpcTaxiServer_address = "127.0.0.1"
const GrpcTaxiServer_port = ":50563"

const GrpcTaxiHardwareServer_address = "127.0.0.1"
const GrpcTaxiHardwareServer_port = ":50652"

const GrpcTaxiHardwareStreamServer_address = "127.0.0.1"
const GrpcTaxiHardwareStreamServer_port = ":50654"

// *** Costs ***
// Taxi ride costs
// Speed: x; 30 öre per second
// Accelaration: Max accelaration = x; 30 öre per second
// Time : 0,5*x; 15 öre per second
//
// Total max: 75 öre per hour [2.700 kr/hour], Constant speed: 45 öre per second [1.620 kr/hour]
//
const SpeedSatoshiPerSecond = 1000
const AccelarationSatoshiPerSecond = 1000
const TimeSatoshiPerSecond = 1000

