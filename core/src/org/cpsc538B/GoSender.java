package org.cpsc538B;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.utils.Disposable;
import com.fasterxml.jackson.databind.JsonNode;
import com.google.common.collect.ImmutableBiMap;
import com.google.common.collect.ImmutableMap;
import lombok.AllArgsConstructor;
import lombok.Data;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.util.*;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.SynchronousQueue;

/**
 * Created by newmanne on 14/03/15.
 */
public class GoSender implements Disposable {

    private Process goProcess;
    private final Queue<Object> goEvents = new SynchronousQueue<>();
    private InetAddress goAddress;
    private DatagramSocket serverSocket;
    private int goPort;

    void init(final String masterAddress) {
        // spawn server
        new Thread(new Runnable() {
            @Override
            public void run() {
                final DatagramSocket serverSocket;
                try {
                    serverSocket = new DatagramSocket(0);
                } catch (SocketException e) {
                    e.printStackTrace();
                    throw new RuntimeException("Couldn't make server", e);
                }
                Gdx.app.log(TronP2PGame.SERVER_TAG, "UDP server started on port " + serverSocket.getLocalPort());

                // spawn go
                // stuff runs from core/assets
                new Thread(new Runnable() {
                    @Override
                    public void run() {
                        try {

                            Runtime r = Runtime.getRuntime();
                            goProcess = new ProcessBuilder("go", "run", "../../go/server.go", Integer.toString(serverSocket.getLocalPort()), masterAddress).start();

                            BufferedReader stdInput = new BufferedReader(new InputStreamReader(goProcess.getInputStream()));
                            BufferedReader stdError = new BufferedReader(new InputStreamReader(goProcess.getErrorStream()));
                            // read the output from the command
                            System.out.println("Here is the standard output of the command:\n");
                            String s;
                            while ((s = stdInput.readLine()) != null) {
                                System.out.println(s);
                            }
                            // read any errors from the attempted command
                            System.out.println("Here is the standard error of the command (if any):\n");
                            while ((s = stdError.readLine()) != null) {
                                System.out.println(s);
                            }
                        } catch (IOException e) {
                            e.printStackTrace();
                        }
                    }
                }).start();

                // server stuff
                byte[] receiveData = new byte[2048];

                while (true) {
                    DatagramPacket receivePacket = new DatagramPacket(receiveData, receiveData.length);
                    try {
                        serverSocket.receive(receivePacket);
                    } catch (IOException e) {
                        e.printStackTrace();
                    }
                    String sentence = new String(receivePacket.getData()).trim();
                    System.out.println("RECEIVED: " + sentence);
                    try {
                        final JsonNode jsonNode = JSONUtils.getMapper().readTree(sentence);
                        final String name = jsonNode.get("name").asText();
                        final Object event = JSONUtils.toObject(jsonNode.get("event").asText(), nameToEvent.get(name));
                    } catch (IOException e) {
                        e.printStackTrace();
                    }
                }
            }
        }).start();

    }

    public void sendToGo(Object event) {
        final String jsonString = JSONUtils.toString(new GoEvent(nameToEvent.inverse().get(event.getClass()), event));
        byte[] sendData = jsonString.getBytes();
        DatagramPacket sendPacket = new DatagramPacket(sendData, sendData.length, goAddress, goPort);
        try {
            serverSocket.send(sendPacket);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public Collection<Object> getGoEvents() {
        List<Object> events = new ArrayList<>();
        while (goEvents.isEmpty()) {
            events.add(goEvents.poll());
        }
        return events;
    }

    @Override
    public void dispose() {
        goProcess.destroyForcibly();
    }

    final ImmutableBiMap<String, Class<?>> nameToEvent = ImmutableBiMap.of("roundStart", RoundStartEvent.class, "myMove", MoveEvent.class, "moves", MovesEvent.class);

    @Data
    @AllArgsConstructor
    public static class GoEvent {
        String name;
        Object event;
    }

    @Data
    public static class RoundStartEvent {
        int round;
        int pid;
    }

    @Data
    public static class MoveEvent {
        public MoveEvent(PositionAndDirection positionAndDirection, int pid) {
            this.positionAndDirection = positionAndDirection;
            this.pid = pid;
        }

        PositionAndDirection positionAndDirection;
        int pid;
    }

    @Data
    public static class MovesEvent {
        List<MoveEvent> moveEvents;
    }

}
