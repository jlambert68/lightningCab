// package: customer_ui_api
// file: examplecom/library/customer_ui_api.proto

var examplecom_library_customer_ui_api_pb = require("../../examplecom/library/customer_ui_api_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var Customer_UI = (function () {
  function Customer_UI() {}
  Customer_UI.serviceName = "customer_ui_api.Customer_UI";
  return Customer_UI;
}());

Customer_UI.AskTaxiForPrice = {
  methodName: "AskTaxiForPrice",
  service: Customer_UI,
  requestStream: false,
  responseStream: false,
  requestType: examplecom_library_customer_ui_api_pb.EmptyParameter,
  responseType: examplecom_library_customer_ui_api_pb.Price_UI
};

Customer_UI.AcceptPrice = {
  methodName: "AcceptPrice",
  service: Customer_UI,
  requestStream: false,
  responseStream: false,
  requestType: examplecom_library_customer_ui_api_pb.EmptyParameter,
  responseType: examplecom_library_customer_ui_api_pb.AckNackResponse
};

Customer_UI.HaltPayments = {
  methodName: "HaltPayments",
  service: Customer_UI,
  requestStream: false,
  responseStream: false,
  requestType: examplecom_library_customer_ui_api_pb.HaltPaymentRequest,
  responseType: examplecom_library_customer_ui_api_pb.AckNackResponse
};

Customer_UI.LeaveTaxi = {
  methodName: "LeaveTaxi",
  service: Customer_UI,
  requestStream: false,
  responseStream: false,
  requestType: examplecom_library_customer_ui_api_pb.EmptyParameter,
  responseType: examplecom_library_customer_ui_api_pb.AckNackResponse
};

Customer_UI.UIPriceAndStateStream = {
  methodName: "UIPriceAndStateStream",
  service: Customer_UI,
  requestStream: false,
  responseStream: true,
  requestType: examplecom_library_customer_ui_api_pb.EmptyParameter,
  responseType: examplecom_library_customer_ui_api_pb.UIPriceAndStateRespons
};

exports.Customer_UI = Customer_UI;

function Customer_UIClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

Customer_UIClient.prototype.askTaxiForPrice = function askTaxiForPrice(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Customer_UI.AskTaxiForPrice, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

Customer_UIClient.prototype.acceptPrice = function acceptPrice(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Customer_UI.AcceptPrice, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

Customer_UIClient.prototype.haltPayments = function haltPayments(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Customer_UI.HaltPayments, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

Customer_UIClient.prototype.leaveTaxi = function leaveTaxi(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Customer_UI.LeaveTaxi, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

Customer_UIClient.prototype.uIPriceAndStateStream = function uIPriceAndStateStream(requestMessage, metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.invoke(Customer_UI.UIPriceAndStateStream, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onMessage: function (responseMessage) {
      listeners.data.forEach(function (handler) {
        handler(responseMessage);
      });
    },
    onEnd: function (status, statusMessage, trailers) {
      listeners.end.forEach(function (handler) {
        handler();
      });
      listeners.status.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners = null;
    }
  });
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

exports.Customer_UIClient = Customer_UIClient;

