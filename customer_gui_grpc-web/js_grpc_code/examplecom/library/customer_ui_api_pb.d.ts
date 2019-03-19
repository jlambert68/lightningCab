// package: customer_ui_api
// file: examplecom/library/customer_ui_api.proto

import * as jspb from "google-protobuf";

export class EmptyParameter extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EmptyParameter.AsObject;
  static toObject(includeInstance: boolean, msg: EmptyParameter): EmptyParameter.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: EmptyParameter, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EmptyParameter;
  static deserializeBinaryFromReader(message: EmptyParameter, reader: jspb.BinaryReader): EmptyParameter;
}

export namespace EmptyParameter {
  export type AsObject = {
  }
}

export class AckNackResponse extends jspb.Message {
  getAcknack(): boolean;
  setAcknack(value: boolean): void;

  getComments(): string;
  setComments(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AckNackResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AckNackResponse): AckNackResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AckNackResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AckNackResponse;
  static deserializeBinaryFromReader(message: AckNackResponse, reader: jspb.BinaryReader): AckNackResponse;
}

export namespace AckNackResponse {
  export type AsObject = {
    acknack: boolean,
    comments: string,
  }
}

export class HaltPaymentRequest extends jspb.Message {
  getHaltpayment(): boolean;
  setHaltpayment(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): HaltPaymentRequest.AsObject;
  static toObject(includeInstance: boolean, msg: HaltPaymentRequest): HaltPaymentRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: HaltPaymentRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): HaltPaymentRequest;
  static deserializeBinaryFromReader(message: HaltPaymentRequest, reader: jspb.BinaryReader): HaltPaymentRequest;
}

export namespace HaltPaymentRequest {
  export type AsObject = {
    haltpayment: boolean,
  }
}

export class Price_UI extends jspb.Message {
  getAcknack(): boolean;
  setAcknack(value: boolean): void;

  getComments(): string;
  setComments(value: string): void;

  getSpeedAmountSatoshi(): number;
  setSpeedAmountSatoshi(value: number): void;

  getAccelerationAmountSatoshi(): number;
  setAccelerationAmountSatoshi(value: number): void;

  getTimeAmountSatoshi(): number;
  setTimeAmountSatoshi(value: number): void;

  getSpeedAmountSek(): number;
  setSpeedAmountSek(value: number): void;

  getAccelerationAmountSek(): number;
  setAccelerationAmountSek(value: number): void;

  getTimeAmountSek(): number;
  setTimeAmountSek(value: number): void;

  getTimeunit(): TimeUnit;
  setTimeunit(value: TimeUnit): void;

  getPaymentrequestinterval(): number;
  setPaymentrequestinterval(value: number): void;

  getPriceunit(): PriceUnit;
  setPriceunit(value: PriceUnit): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Price_UI.AsObject;
  static toObject(includeInstance: boolean, msg: Price_UI): Price_UI.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Price_UI, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Price_UI;
  static deserializeBinaryFromReader(message: Price_UI, reader: jspb.BinaryReader): Price_UI;
}

export namespace Price_UI {
  export type AsObject = {
    acknack: boolean,
    comments: string,
    speedAmountSatoshi: number,
    accelerationAmountSatoshi: number,
    timeAmountSatoshi: number,
    speedAmountSek: number,
    accelerationAmountSek: number,
    timeAmountSek: number,
    timeunit: TimeUnit,
    paymentrequestinterval: number,
    priceunit: PriceUnit,
  }
}

export class RPCMethods extends jspb.Message {
  getAsktaxiforprice(): boolean;
  setAsktaxiforprice(value: boolean): void;

  getAcceptprice(): boolean;
  setAcceptprice(value: boolean): void;

  getHaltpaymentsTrue(): boolean;
  setHaltpaymentsTrue(value: boolean): void;

  getHaltpaymentsFalse(): boolean;
  setHaltpaymentsFalse(value: boolean): void;

  getLeavetaxi(): boolean;
  setLeavetaxi(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RPCMethods.AsObject;
  static toObject(includeInstance: boolean, msg: RPCMethods): RPCMethods.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RPCMethods, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RPCMethods;
  static deserializeBinaryFromReader(message: RPCMethods, reader: jspb.BinaryReader): RPCMethods;
}

export namespace RPCMethods {
  export type AsObject = {
    asktaxiforprice: boolean,
    acceptprice: boolean,
    haltpaymentsTrue: boolean,
    haltpaymentsFalse: boolean,
    leavetaxi: boolean,
  }
}

export class UIPriceAndStateRespons extends jspb.Message {
  getAcknack(): boolean;
  setAcknack(value: boolean): void;

  getComments(): string;
  setComments(value: string): void;

  getSpeedAmountSatoshi(): number;
  setSpeedAmountSatoshi(value: number): void;

  getAccelerationAmountSatoshi(): number;
  setAccelerationAmountSatoshi(value: number): void;

  getTimeAmountSatoshi(): number;
  setTimeAmountSatoshi(value: number): void;

  getSpeedAmountSek(): number;
  setSpeedAmountSek(value: number): void;

  getAccelerationAmountSek(): number;
  setAccelerationAmountSek(value: number): void;

  getTimeAmountSek(): number;
  setTimeAmountSek(value: number): void;

  getTotalAmountSatoshi(): number;
  setTotalAmountSatoshi(value: number): void;

  getTotalAmountSek(): number;
  setTotalAmountSek(value: number): void;

  getTimestamp(): number;
  setTimestamp(value: number): void;

  hasAllowedrpcmethods(): boolean;
  clearAllowedrpcmethods(): void;
  getAllowedrpcmethods(): RPCMethods | undefined;
  setAllowedrpcmethods(value?: RPCMethods): void;

  getCurrentTaxirideSatoshi(): number;
  setCurrentTaxirideSatoshi(value: number): void;

  getCurrentTaxiRideSek(): number;
  setCurrentTaxiRideSek(value: number): void;

  getCurrentWalletbalanceSatoshi(): number;
  setCurrentWalletbalanceSatoshi(value: number): void;

  getCurrentWalletbalanceSek(): number;
  setCurrentWalletbalanceSek(value: number): void;

  getAveragePaymentAmountSatoshi(): number;
  setAveragePaymentAmountSatoshi(value: number): void;

  getAveragePaymentAmountSek(): number;
  setAveragePaymentAmountSek(value: number): void;

  getAvaregeNumberOfPayments(): number;
  setAvaregeNumberOfPayments(value: number): void;

  getAccelarationPercent(): number;
  setAccelarationPercent(value: number): void;

  getSpeedPercent(): number;
  setSpeedPercent(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UIPriceAndStateRespons.AsObject;
  static toObject(includeInstance: boolean, msg: UIPriceAndStateRespons): UIPriceAndStateRespons.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: UIPriceAndStateRespons, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UIPriceAndStateRespons;
  static deserializeBinaryFromReader(message: UIPriceAndStateRespons, reader: jspb.BinaryReader): UIPriceAndStateRespons;
}

export namespace UIPriceAndStateRespons {
  export type AsObject = {
    acknack: boolean,
    comments: string,
    speedAmountSatoshi: number,
    accelerationAmountSatoshi: number,
    timeAmountSatoshi: number,
    speedAmountSek: number,
    accelerationAmountSek: number,
    timeAmountSek: number,
    totalAmountSatoshi: number,
    totalAmountSek: number,
    timestamp: number,
    allowedrpcmethods?: RPCMethods.AsObject,
    currentTaxirideSatoshi: number,
    currentTaxiRideSek: number,
    currentWalletbalanceSatoshi: number,
    currentWalletbalanceSek: number,
    averagePaymentAmountSatoshi: number,
    averagePaymentAmountSek: number,
    avaregeNumberOfPayments: number,
    accelarationPercent: number,
    speedPercent: number,
  }
}

export enum PriceUnit {
  SATOSHIPERSECOND = 0,
}

export enum TimeUnit {
  SECONDSBETWEENPAYMENTMENTREQUESTS = 0,
  MILLISECONDSBETWEENPAYMENTMENTREQUESTS = 1,
}

