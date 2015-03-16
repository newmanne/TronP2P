package org.cpsc538B.go;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.utils.Disposable;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.JsonToken;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.google.common.collect.ImmutableBiMap;
import com.google.common.collect.ImmutableMap;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.cpsc538B.model.Direction;
import org.cpsc538B.model.PositionAndDirection;
import org.cpsc538B.TronP2PGame;
import org.cpsc538B.utils.JSONUtils;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.util.*;
import java.util.concurrent.ArrayBlockingQueue;

/**
 * Created by newmanne on 14/03/15.
 */
public class GoSender implements Disposable {

    private Process goProcess;
    private final Queue<Object> goEvents = new ArrayBlockingQueue<>(20);
    private InetAddress goAddress;
    private DatagramSocket serverSocket;
    private int goPort;
    final ImmutableMap<String, Class<?>> nameToEvent = ImmutableMap.of("roundStart", RoundStartEvent.class, "myMove", MoveEvent.class, "moves", MovesEvent.class);

    public void init(final String masterAddress, final boolean leader) {
        // spawn server
        try {
            serverSocket = new DatagramSocket(0);
        } catch (SocketException e) {
            e.printStackTrace();
            throw new RuntimeException("Couldn't make server", e);
        }
        Gdx.app.log(TronP2PGame.SERVER_TAG, "UDP server started on port " + serverSocket.getLocalPort());

        // spawn go
        // stuff runs from core/assets
        new Thread(() -> {
            try {
                Runtime r = Runtime.getRuntime();
                final ProcessBuilder processBuilder = new ProcessBuilder("go", "run", "../../go/server.go", Integer.toString(serverSocket.getLocalPort()), masterAddress, Boolean.toString(leader));
                Gdx.app.log(TronP2PGame.LOG_TAG, "Running the following command:" + System.lineSeparator() + processBuilder.command() + System.lineSeparator());
                goProcess = processBuilder.start();

                BufferedReader stdInput = new BufferedReader(new InputStreamReader(goProcess.getInputStream()));
                BufferedReader stdError = new BufferedReader(new InputStreamReader(goProcess.getErrorStream()));
                String stdout = null;
                String stderr = null;
                while (true) {
                    while ((stdout = stdInput.readLine()) != null) {
                        Gdx.app.log(TronP2PGame.GO_STDOUT_TAG, stdout);
                    }
                    while ((stderr = stdError.readLine()) != null) {
                        Gdx.app.log(TronP2PGame.GO_STDERR_TAG, stderr);
                    }
                }
            } catch (IOException e) {
                e.printStackTrace();
            }
        }).start();

        // server stuff
        new Thread(() -> {
            byte[] receiveData = new byte[2048];

            while (true) {
                DatagramPacket receivePacket = new DatagramPacket(receiveData, receiveData.length);
                try {
                    serverSocket.receive(receivePacket);
                } catch (IOException e) {
                    e.printStackTrace();
                }
                String sentence = new String(receivePacket.getData()).trim();
                Gdx.app.log(TronP2PGame.SERVER_TAG, "RECEIVED: " + sentence);
                try {
                    goPort = receivePacket.getPort();
                    goAddress = receivePacket.getAddress();
                    final JsonNode jsonNode = JSONUtils.getMapper().readTree(sentence);
                    final String name = jsonNode.get("eventName").asText();
                    Gdx.app.log(TronP2PGame.SERVER_TAG, "Event recieved is of type " + name);
                    final Object event = JSONUtils.getMapper().treeToValue(jsonNode.get(name), nameToEvent.get(name));
                    goEvents.add(event);
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        }).start();
    }

    public void sendToGo(Object event) {
        final String jsonString = JSONUtils.toString(event);
        Gdx.app.log(TronP2PGame.SERVER_TAG, "Sending message " + System.lineSeparator() + jsonString);
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
        while (!goEvents.isEmpty()) {
            events.add(goEvents.poll());
        }
        return events;
    }

    @Override
    public void dispose() {
        goProcess.destroyForcibly();
    }


    @Data
    public static class RoundStartEvent {
        String eventName = "roundStart";
        int pid;
        int round;
    }

    @Data
    @NoArgsConstructor
    public static class MoveEvent {
        String eventName = "myMove";

        public MoveEvent(PositionAndDirection positionAndDirection, int pid) {
            this.x = positionAndDirection.getX();
            this.y = positionAndDirection.getY();
            this.direction = positionAndDirection.getDirection();
            this.pid = pid;
        }

        int x;
        int y;
        Direction direction;

        @JsonIgnore
        public PositionAndDirection getPositionAndDirection() {
            return new PositionAndDirection(x, y, direction);
        }

        int pid;
    }

    @Data
    @JsonDeserialize(using = MovesEventDeserializer.class)
    public static class MovesEvent {
        String eventName = "moves";
        List<MoveEvent> moves;
    }

    public static class MovesEventDeserializer extends JsonDeserializer<MovesEvent> {

        @Override
        public MovesEvent deserialize(JsonParser jsonParser, DeserializationContext ctxt) throws IOException, JsonProcessingException {
            if (jsonParser.getCurrentToken() == JsonToken.START_ARRAY) {
                List<MoveEvent> permissions = new ArrayList<>();
                while (jsonParser.nextToken() != JsonToken.END_ARRAY) {
                    permissions.add(jsonParser.readValueAs(MoveEvent.class));
                }
                MovesEvent movesEvent = new MovesEvent();
                movesEvent.setMoves(permissions);
                return movesEvent;
            }
            throw new IllegalStateException();
        }
    }

}