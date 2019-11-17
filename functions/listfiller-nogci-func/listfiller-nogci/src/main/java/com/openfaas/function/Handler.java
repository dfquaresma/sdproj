package com.openfaas.function;

import com.openfaas.model.IHandler;
import com.openfaas.model.IResponse;
import com.openfaas.model.IRequest;
import com.openfaas.model.Response;

import java.util.List;
import java.util.ArrayList;

public class Handler implements com.openfaas.model.IHandler {

    public IResponse Handle(IRequest req) {
        callFunction();
        Response res = new Response();
	    res.setBody("ListFiller NOGCI! :|");
	    return res;
    }

    public void callFunction() {
        int size = (int) Math.pow(2, 15);
        List<Integer> list  = new ArrayList();
        for (int i = 0; i < size; i++) {
                list.add(i);
        }
    }
}
