// package: customer_ui_api
// file: examplecom/library/customer_ui_api.proto

import * as examplecom_library_customer_ui_api_pb from "../../examplecom/library/customer_ui_api_pb";
import {grpc} from "@improbable-eng/grpc-web";

type Customer_UIAskTaxiForPrice = {
  readonly methodName: string;
  readonly service: typeof Customer_UI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof examplecom_library_customer_ui_api_pb.EmptyParameter;
  readonly responseType: typeof examplecom_library_customer_ui_api_pb.Price_UI;
};

type Customer_UIAcceptPrice = {
  readonly methodName: string;
  readonly service: typeof Customer_UI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof examplecom_library_customer_ui_api_pb.EmptyParameter;
  readonly responseType: typeof examplecom_library_customer_ui_api_pb.AckNackResponse;
};

type Customer_UIHaltPayments = {
  readonly methodName: string;
  readonly service: typeof Customer_UI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof examplecom_library_customer_ui_api_pb.HaltPaymentRequest;
  readonly responseType: typeof examplecom_library_customer_ui_api_pb.AckNackResponse;
};

type Customer_UILeaveTaxi = {
  readonly methodName: string;
  readonly service: typeof Customer_UI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof examplecom_library_customer_ui_api_pb.EmptyParameter;
  readonly responseType: typeof examplecom_library_customer_ui_api_pb.AckNackResponse;
};

type Customer_UIUIPriceAndStateStream = {
  readonly methodName: string;
  readonly service: typeof Customer_UI;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof examplecom_library_customer_ui_api_pb.EmptyParameter;
  readonly responseType: typeof examplecom_library_customer_ui_api_pb.UIPriceAndStateRespons;
};

export class Customer_UI {
  static readonly serviceName: string;
  static readonly AskTaxiForPrice: Customer_UIAskTaxiForPrice;
  static readonly AcceptPrice: Customer_UIAcceptPrice;
  static readonly HaltPayments: Customer_UIHaltPayments;
  static readonly LeaveTaxi: Customer_UILeaveTaxi;
  static readonly UIPriceAndStateStream: Customer_UIUIPriceAndStateStream;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: () => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: () => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: () => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class Customer_UIClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  askTaxiForPrice(
    requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.Price_UI|null) => void
  ): UnaryResponse;
  askTaxiForPrice(
    requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.Price_UI|null) => void
  ): UnaryResponse;
  acceptPrice(
    requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.AckNackResponse|null) => void
  ): UnaryResponse;
  acceptPrice(
    requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.AckNackResponse|null) => void
  ): UnaryResponse;
  haltPayments(
    requestMessage: examplecom_library_customer_ui_api_pb.HaltPaymentRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.AckNackResponse|null) => void
  ): UnaryResponse;
  haltPayments(
    requestMessage: examplecom_library_customer_ui_api_pb.HaltPaymentRequest,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.AckNackResponse|null) => void
  ): UnaryResponse;
  leaveTaxi(
    requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.AckNackResponse|null) => void
  ): UnaryResponse;
  leaveTaxi(
    requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter,
    callback: (error: ServiceError|null, responseMessage: examplecom_library_customer_ui_api_pb.AckNackResponse|null) => void
  ): UnaryResponse;
  uIPriceAndStateStream(requestMessage: examplecom_library_customer_ui_api_pb.EmptyParameter, metadata?: grpc.Metadata): ResponseStream<examplecom_library_customer_ui_api_pb.UIPriceAndStateRespons>;
}

