package main

import (
	"log"
	"golang.org/x/net/context"
	"jlambert/lightningCab/customer_server/customer_grpc_ui_api"
	"jlambert/lightningCab/common_config"
)

// ************************************************************************************
// UI-customer Ask for Price

func (s *Customer_UIServiceServer) AskTaxiForPrice(ctx context.Context, emptyParameter *customer_ui_api.EmptyParameter) (*customer_ui_api.Price_UI, error) {

	log.Println("Incoming: 'AskTaxiForPrice'")

	returnMessage := &customer_ui_api.Price_UI{
		false,
		"",
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
	}

	// Check if State machine accepts State change
	err := customer.askTaxiForPrice(true)

	if err == nil {

		switch  err.(type) {
		case nil:
			err = customer.acceptPrice(false)
			if err != nil {
				logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"

			} else {

				logMessagesWithOutError(4, "Success in change if state: ")

				returnMessage = &customer_ui_api.Price_UI{
					true,
					customer.lastRecievedPriceInfo.Comments,
					customer.lastRecievedPriceInfo.GetSpeed(),
					customer.lastRecievedPriceInfo.GetAcceleration(),
					customer.lastRecievedPriceInfo.GetTime(),
					float32(customer.lastRecievedPriceInfo.GetSpeed()) * common_config.BTCSEK,
					float32(customer.lastRecievedPriceInfo.GetAcceleration()) * common_config.BTCSEK,
					float32(customer.lastRecievedPriceInfo.GetTime()) * common_config.BTCSEK,
					0, //customer.lastRecievedPriceInfo.Timeunit,
					0, //customer.lastRecievedPriceInfo.PaymentRequestInterval,
					0, //customer.lastRecievedPriceInfo.Priceunit,
				}
			}

		default:
			logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"

		}
	} else {
		logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
		returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"
	}
	return returnMessage, nil
}

// ************************************************************************************
// UI-customer accepts price

func (s *Customer_UIServiceServer) AcceptPrice(ctx context.Context, emptyParameter *customer_ui_api.EmptyParameter) (*customer_ui_api.AckNackResponse, error) {

	log.Println("Incoming: 'AcceptPrice'")

	returnMessage := &customer_ui_api.AckNackResponse{
		false,
		"",
	}

	// Check if State machine accepts State change
	err := customer.askTaxiForPrice(true)

	if err == nil {

		switch  err.(type) {
		case nil:
			err = customer.acceptPrice(false)
			if err != nil {
				logMessagesWithError(4, "State machine is not in correct state to be able have customer accept price: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able have customer accept price"

			} else {

				logMessagesWithOutError(4, "Success in change if state: ")

				returnMessage = &customer_ui_api.AckNackResponse{
					true,
					customer.lastRecievedPriceAccept.Comments,
				}
			}

		default:
			logMessagesWithError(4, "State machine is not in correct state to be able have customer accept price: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able have customer accept price"

		}
	} else {
		logMessagesWithError(4, "State machine is not in correct state to be able have customer accept price: ", err)
		returnMessage.Comments = "State machine is not in correct state to be able have customer accept price"
	}
	return returnMessage, nil
}

// ************************************************************************************
// UI-customer halts payments

func (s *Customer_UIServiceServer) HaltPayments(ctx context.Context, haltPaymentRequestMessage *customer_ui_api.HaltPaymentRequest) (*customer_ui_api.AckNackResponse, error) {

	log.Println("Incoming: 'HaltPayments' with parameter: ", haltPaymentRequestMessage.Haltpayment)

	returnMessage := &customer_ui_api.AckNackResponse{
		false,
		"",
	}

	// Decide of user wants to Halt or un-Halt payment
	switch haltPaymentRequestMessage.Haltpayment {

	case true: //Halt Payments
		// Check if State machine accepts State change
		err := customer.askTaxiForPrice(true)

		if err == nil {

			switch  err.(type) {
			case nil:
				err = customer.acceptPrice(false)
				if err != nil {
					logMessagesWithError(4, "State machine is not in correct state to be able to halt payment: ", err)
					returnMessage.Comments = "State machine is not in correct state to be able to halt payment"

				} else {

					logMessagesWithOutError(4, "Success in change if state: ")

					returnMessage = &customer_ui_api.AckNackResponse{
						true,
						"Success in change of state and Halt payments",
					}
				}

			default:
				logMessagesWithError(4, "State machine is not in correct state to be able to halt payment: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able to halt payment"

			}
		} else {
			logMessagesWithError(4, "State machine is not in correct state to be able to halt payment: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able to halt payment"
		}

	case false: //Unhalt payments
		if customer.CustomerStateMachine.IsInState(StateCustomerHaltedPayments) == true {

			// Correct state for Un-Halt payments
			switch customer.stateBeforeHaltPayments {

			// Go back to wait for PaymentRequest
			case StateCustomerWaitingForPaymentRequest:
				currentTrigger := TriggerCustomerContiniueToWaitForPaymentRequest

				err := customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
				if err != nil {
					logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)
					returnMessage.Comments = "State machine is not in correct state to be able to un-halt payment"
				} else {
					returnMessage = &customer_ui_api.AckNackResponse{
						true,
						"Success in change of state and Un-Halt payments",
					}
				}

				// Go back and pay paymentRequest
			case StateCustomerPaymentRequestReceived:
				currentTrigger := TriggerCustomerWillContinueToPay

				err := customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
				if err != nil {
					logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)
					returnMessage.Comments = "State machine is not in correct state to be able to un-halt payment"
				} else {
					returnMessage = &customer_ui_api.AckNackResponse{
						true,
						"Success in change of state and Un-Halt payments",
					}
				}
			}

		} else {

			// Not Correct state to be able of un-halting payments
			logMessagesWithOutError(4, "State machine is not in correct state to be able to un-halt payment")
			returnMessage.Comments = "State machine is not in correct state to be able to un-halt payment"
		}

	}
	return returnMessage, nil
}

// ************************************************************************************
// UI-Customer leaves Taxi

func (s *Customer_UIServiceServer) LeaveTaxi(ctx context.Context, emptyParameter *customer_ui_api.EmptyParameter) (*customer_ui_api.AckNackResponse, error) {

	log.Println("Incoming: 'LeaveTaxi'")

	returnMessage := &customer_ui_api.AckNackResponse{
		false,
		"",
	}

	// Check if State machine accepts State change
	err := customer.leaveTaxi(true)

	if err == nil {

		switch  err.(type) {
		case nil:
			err = customer.leaveTaxi(false)
			if err != nil {
				logMessagesWithError(4, "State machine is not in correct state to be able have customer leave taxi: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able have customer leave taxi"

			} else {

				logMessagesWithOutError(4, "Success in change if state: ")

				returnMessage = &customer_ui_api.AckNackResponse{
					true,
					"State machine is in correct state and customer left taxi",
				}
			}

		default:
			logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"

		}
	} else {
		logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
		returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"
	}
	return returnMessage, nil
}