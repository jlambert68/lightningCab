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
const usdsek = 8.88            //SEK per USD
const btcusd = 5890            //USD per BTC
const btcsek = btcusd * usdsek //SEK per BTC

const SpeedSEKPerSecond = 0.30                              //SEK
const MaxAccelarationSEKPerSecond = SpeedSEKPerSecond * 0.5 //SEK
const TimeSEKPerSecond = SpeedSEKPerSecond                  //SEK

const SatoshisPerBTC = 1000000000

const SpeedSatoshiPerSecond = SpeedSEKPerSecond / btcsek * SatoshisPerBTC
const MaxAccelarationSatoshiPerSecond = MaxAccelarationSEKPerSecond / btcsek * SatoshisPerBTC
const TimeSatoshiPerSecond = TimeSEKPerSecond / btcsek * SatoshisPerBTC

const MilliSecondsBetweenPaymentRequest = 1000
const SecondsBeforeFirstPaymentTimeOut = 2
const SecondsBeforeSecondPaymentTimeOut = 10


// Simnet or Testnet

const UseSimnet = true

/*
const defaultLndTollGRPCHost = "localhost:10001"
const defaultLndTaxiGRPCHost = "localhost:10002"
const defaultLndCustomerGRPCHost = "localhost:10003"
*/

