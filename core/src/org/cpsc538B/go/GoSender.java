package org.cpsc538B.go;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.utils.Disposable;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.JsonNode;
import com.google.common.base.Preconditions;
import com.google.common.collect.ImmutableMap;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.cpsc538B.TronP2PGame;
import org.cpsc538B.model.Direction;
import org.cpsc538B.model.PositionAndDirection;
import org.cpsc538B.screens.GameScreen;
import org.cpsc538B.utils.JSONUtils;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.*;
import java.util.concurrent.ArrayBlockingQueue;

/**
 * Created by newmanne on 14/03/15.
 */
public class GoSender implements Disposable {

    private Process goProcess;
    private final Queue<Object> goEvents = new ArrayBlockingQueue<>(20);
    private InetAddress goAddress;
    private ServerSocket serverSocket;
    private Socket goSocket;
    private int goPort;
    final ImmutableMap<String, Class<?>> nameToEvent = ImmutableMap.<String, Class<?>>builder()
            .put("roundStart", RoundStartEvent.class)
            .put("myMove", MoveEvent.class)
            .put("moves", MovesEvent.class)
            .put("gameStart", GameStartEvent.class)
            .put("gameOver", GameOverEvent.class)
            .build();
    private BufferedReader goInputStream;
    private PrintWriter goOutputStream;

    public void init(final String masterAddress, final String nickname, final boolean leader, GoInitializedCallback callback) {
        // spawn server
        try {
            serverSocket = new ServerSocket(0);
        } catch (IOException e) {
            e.printStackTrace();
            throw new RuntimeException("Couldn't make server", e);
        }
        Gdx.app.log(TronP2PGame.SERVER_TAG, "Java listener server started on port " + serverSocket.getLocalPort());

        // spawn go
        // stuff runs from core/assets
        new Thread(() -> {
            try {
                Runtime r = Runtime.getRuntime();
                final ProcessBuilder processBuilder = new ProcessBuilder("go", "run", "../../go/server.go", Integer.toString(serverSocket.getLocalPort()), masterAddress, Boolean.toString(leader), Integer.toString(GameScreen.GRID_WIDTH), Integer.toString(GameScreen.GRID_HEIGHT), nickname);
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
            try {
                goSocket = serverSocket.accept();
            } catch (IOException e) {
                Gdx.app.log(TronP2PGame.SERVER_TAG, "Failed to accept go client", e);
            }
            Gdx.app.log(TronP2PGame.SERVER_TAG, "Go client connected from " + goSocket.toString());
            try {
                goInputStream = new BufferedReader(new InputStreamReader(goSocket.getInputStream()));
                goOutputStream = new PrintWriter(goSocket.getOutputStream(), true);
            } catch (IOException e) {
                e.printStackTrace();
            }
            while (goSocket.isConnected()) {
                try {
                    // Note that all messages are delineated by lines
                    String message = goInputStream.readLine();
                    Gdx.app.log(TronP2PGame.SERVER_TAG, "RECEIVED: " + message);

                    final JsonNode jsonNode = JSONUtils.getMapper().readTree(message);
                    final String name = jsonNode.get("eventName").asText();
                    final int round = jsonNode.get("round").asInt();
                    Gdx.app.log(TronP2PGame.SERVER_TAG, "Event received is of type " + name + " for round " + round);
                    final Object event = JSONUtils.getMapper().treeToValue(jsonNode.get(name), nameToEvent.get(name));
                    // special case if a game start event is received
                    if (event instanceof GameStartEvent) {
                        GameStartEvent gameStartEvent = (GameStartEvent) event;
                        callback.onGameStarted(gameStartEvent.getPid(), gameStartEvent.getStartingPositions(), gameStartEvent.getNicknames());
                    } else {
                        goEvents.add(event);
                    }
                } catch (IOException e) {
                    e.printStackTrace();
                    break;
                }
            }
        }).start();
    }

    public void sendToGo(Object event) {
        Preconditions.checkNotNull(goOutputStream != null, "Go output stream is null");
        Preconditions.checkNotNull(event, "Event is null");
        final String jsonString = JSONUtils.toString(event);
        Gdx.app.log(TronP2PGame.SERVER_TAG, "Sending message: " + jsonString);
        goOutputStream.println(jsonString);
    }

    public void sendToGo(String string) {
        Preconditions.checkNotNull(goOutputStream != null, "Go output stream is null");
        goOutputStream.println(string);
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
        if (goProcess != null) {
            goProcess.destroyForcibly();
        }
        if (goOutputStream != null) {
            goOutputStream.close();
        }
        try {
            if (goInputStream != null) {
                goInputStream.close();
            }
            if (goSocket != null) {
                goSocket.close();
            }
            if (serverSocket != null) {
                serverSocket.close();
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }


    @Data
    public static class RoundStartEvent {
        int round;
    }

    @Data
    public static class GameStartEvent {
        String pid;
        Map<String, PositionAndDirection> startingPositions;
        Map<String, String> nicknames;
	Map<String, String> addresses;
    }


    @Data
    @NoArgsConstructor
    public static class MoveEvent {
        String eventName = "myMove";

        public MoveEvent(Direction direction, String pid, int round) {
            this.direction = direction;
            this.pid = pid;
            this.round = round;
        }

        Direction direction;
        String pid;
        int round;
    }

    @Data
    public static class MovesEvent {
        List<Map<String, PositionAndDirection>> moves;
        int round;
    }

    public interface GoInitializedCallback {
        void onGameStarted(String pid, Map<String, PositionAndDirection> startingPositions, Map<String, String> nicknames);
    }

    @Data
    @NoArgsConstructor
    public static class GameOverEvent {
        List<String> pidsInOrderOfDeath;
    }
}
