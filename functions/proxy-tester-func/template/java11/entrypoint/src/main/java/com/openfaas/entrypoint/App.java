// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package com.openfaas.entrypoint;

import com.openfaas.model.*;

import java.net.*;
import java.io.*;
import java.util.*;

public class App { 
    
    public static void main(String[] args) throws Exception {
        runServerWithNoDependencies();
    }

    // Look ma, no... dependencies
    private static void runServerWithNoDependencies() {
        int port = Integer.parseInt(System.getenv("entrypoint_port"));
        String newLine="\r\n";
        IResponse res = new Response(); // we ignore body and headers at first moment
        IHandler handler = new com.openfaas.function.Handler();
        try { 
            ServerSocket socket = new ServerSocket(port);
            while (true) {
                Socket connection = socket.accept();
                try { 
                    BufferedReader in = new BufferedReader(new InputStreamReader(connection.getInputStream()));
                    OutputStream out = new BufferedOutputStream(connection.getOutputStream());
                    PrintStream pout = new PrintStream(out);

                // read first line of request
                String request = in.readLine();
                if (request == null) continue;

                // we ignore the rest
                while (true) { 
                    String ignore = in.readLine();
                    if (ignore == null || ignore.length() == 0) break;
                }

                if (!request.startsWith("GET ") || !(request.endsWith(" HTTP/1.0") || request.endsWith(" HTTP/1.1"))) {
                    // bad request
                    pout.print("HTTP/1.0 400 Bad Request" + newLine + newLine);
                } else {
                    res = handler.Handle(null);
                    String status = "200 OK";
                    if (res.getStatusCode() != 200) {
                        status = "503 SERVICE UNAVAILABLE";
                    }
                    pout.print(
                        "HTTP/1.0 " + status + newLine +
                        "Content-Type: text/plain" + newLine +
                        "Date: " + new Date() + newLine +
                        "Content-length: " + res.getBody().length() + newLine + newLine +
                        res.getBody());
                }

                pout.close();
                } catch (Throwable tri) {System.err.println("Error handling request: " + tri);}
            }
        } catch (Throwable tr) {System.err.println("Could not start server: " + tr);}
    }
}
