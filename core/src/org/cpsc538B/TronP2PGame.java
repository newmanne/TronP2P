package org.cpsc538B;

import com.badlogic.gdx.ApplicationAdapter;
import com.badlogic.gdx.Game;
import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.graphics.GL20;
import com.badlogic.gdx.graphics.Texture;
import com.badlogic.gdx.graphics.g2d.SpriteBatch;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import lombok.Getter;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;

public class TronP2PGame extends Game {

    @Getter
    private SpriteBatch spritebatch;
    @Getter
    private ShapeRenderer shapeRenderer;
    @Getter
    private Assets assets;
    @Getter
    private StartScreen startScreen;

    private Process goProcess;

    public final static String LOG_TAG = "TRON";
    public final static String SERVER_TAG = "SERVER";


    @Override
    public void create() {
        Gdx.app.log(LOG_TAG, "Starting game!");
        spritebatch = new SpriteBatch();
        shapeRenderer = new ShapeRenderer();
        assets = new Assets();
        startScreen = new StartScreen(this);

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
                Gdx.app.log(SERVER_TAG, "UDP server started on port " + serverSocket.getLocalPort());
                // spawn go
                // stuff runs from core/assets
                new Thread(new Runnable() {
                    @Override
                    public void run() {
                        try {

                            Runtime r = Runtime.getRuntime();
                            goProcess = new ProcessBuilder("go", "run", "../../go/server.go", Integer.toString(serverSocket.getLocalPort())).start();

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

                byte[] receiveData = new byte[1024];
                byte[] sendData = new byte[1024];
                while (true) {
                    DatagramPacket receivePacket = new DatagramPacket(receiveData, receiveData.length);
                    try {
                        serverSocket.receive(receivePacket);
                    } catch (IOException e) {
                        e.printStackTrace();
                    }
                    String sentence = new String(receivePacket.getData()).trim();
                    System.out.println("RECEIVED: " + sentence);
                    InetAddress IPAddress = receivePacket.getAddress();
                    int port = receivePacket.getPort();
                    String capitalizedSentence = sentence.toUpperCase();
                    sendData = capitalizedSentence.getBytes();
                    DatagramPacket sendPacket = new DatagramPacket(sendData, sendData.length, IPAddress, port);
                    try {
                        serverSocket.send(sendPacket);
                    } catch (IOException e) {
                        e.printStackTrace();
                    }
                }

            }
        }).start();

        setScreen(startScreen);
    }

    @Override
    public void dispose() {
        assets.dispose();
        shapeRenderer.dispose();
        spritebatch.dispose();
        goProcess.destroyForcibly();
    }

}
