package org.cpsc538B;

/**
 * Created by newmanne on 12/03/15.
 */

import com.badlogic.gdx.Application.ApplicationType;
import com.badlogic.gdx.Game;
import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.Input.Orientation;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.graphics.Texture;
import com.badlogic.gdx.graphics.g2d.SpriteBatch;
import com.badlogic.gdx.graphics.g2d.TextureRegion;
import com.badlogic.gdx.graphics.g2d.freetype.FreeTypeFontGenerator;
import com.badlogic.gdx.scenes.scene2d.Actor;
import com.badlogic.gdx.scenes.scene2d.Stage;
import com.badlogic.gdx.scenes.scene2d.ui.*;
import com.badlogic.gdx.scenes.scene2d.utils.ChangeListener;
import com.badlogic.gdx.utils.viewport.StretchViewport;
import com.google.common.collect.ImmutableList;

import java.net.MalformedURLException;
import java.util.List;

public class ClientConnectScreen extends ScreenAdapter {

//    private static final int LABEL_FIELD_PADDING = 20;
//    private static final int FIELD_WIDTH = 400;
//    private static final int defaultFontSize = 25;
//    protected final Game game;
//    private final Stage stage;
//    private final TextField ipAddressField;
//    private final TextField portField;
//    private final TextField nicknameField;
//    private final TextButton connectButton;
//    private final TextButton gameStart;
//    private final TextButton updateButton;
//    private Table table;
//    private final Label waitingText;
//
//    private float VIRTUAL_WIDTH = 800;
//    private float VIRTUAL_HEIGHT = 600;
//    private final FreeTypeFontGenerator fontGenerator;
//    private final Label ipAddressLabel;
//    private final Label portLabel;
//    private final Label nicknameLabel;
//    private final Label announcementLabel;
//
//    private Image backgroundImage;
//
//    public ClientConnectScreen(final TronP2PGame game) {
//        this.game = game;
//        this.stage = new Stage(new StretchViewport(VIRTUAL_WIDTH, VIRTUAL_HEIGHT), game.getSpritebatch());
//
//        final String defaultIP = "localhost";
//        ipAddressField = new TextField(defaultIP, skin);
//        ipAddressField.setMessageText("IP Address");
//
//        portField = new TextField("8080", skin);
//        portField.setMessageText("Port");
//        final List<String> sampleNames = ImmutableList.of("Blinky", "Pacman", "Robocop", "DemonSlayer", "HAL", "ChickenLittle", "HansSolo", "Yoshi", "EcologyFan", "Ghost", "GoLeafsGo", "Batman");
//        final String defaultName = sampleNames.get(RandomUtils.nextInt(0, sampleNames.size()));
//        nicknameField = new TextField(defaultName + RandomStringUtils.randomNumeric(3), skin);
//        nicknameField.setMessageText("Blinky");
//
//        connectButton = new TextButton("Connect", skin);
//        connectButton.addListener(new ChangeListener() {
//            @Override
//            public void changed(ChangeEvent event, Actor actor) {
//                connectButton.setDisabled(true);
//                connectButton.setVisible(false);
//                if (!getSocketIO().isConnected()) {
//                    connect();
//                }
//            }
//        });
//        updateButton = new TextButton("Update Nickname", skin);
//        updateButton.addListener(new ChangeListener() {
//
//            @Override
//            public void changed(ChangeEvent event, Actor actor) {
//                getSocketIO().setNickname(nicknameField.getText());
//                updateButton.setVisible(false);
//                updateButton.setDisabled(true);
//            }
//
//        });
//        updateButton.setDisabled(true);
//        updateButton.setVisible(false);
//
//        gameStart = new TextButton("Start", skin);
//        gameStart.addListener(new ChangeListener() {
//
//            @Override
//            public void changed(ChangeEvent event, Actor actor) {
//                // this is a special event, emit directly to server
//                socketIO.getClient().emit(CommonSocketIOEvents.GAME_START);
//            }
//
//        });
//        gameStart.setVisible(false);
//        gameStart.setDisabled(true);
//
//        ipAddressLabel = new Label("IP Address", skin);
//        portLabel = new Label("Port", skin);
//        nicknameLabel = new Label("Nickname", skin);
//        announcementLabel = new Label("", skin);
//
//        waitingText = new Label("Waiting for host to select the game", skin);
//        waitingText.setVisible(false);
//        buildBackground(skin);
//
//        buildTable(skin);
//        stage.addActor(table);
//    }
//
//    protected void registerEvents() {
//        registerEvent(CommonSocketIOEvents.INVALID_NICKNAME, new EventCallback() {
//
//            @Override
//            public void onEvent(IOAcknowledge ack, Object... args) {
//                new Dialog("Invalid nickname", skin).text("Please pick a different nickname").button("OK").show(stage);
//                updateButton.setDisabled(false);
//                updateButton.setVisible(true);
//                nicknameField.setDisabled(false);
//            }
//        });
//        registerEvent(CommonSocketIOEvents.ELECTED_CLIENT, new EventCallback() {
//
//            @Override
//            public void onEvent(IOAcknowledge ack, Object... args) {
//                getSocketIO().setHost(false);
//                waitingText.setVisible(true);
//            }
//
//        });
//        registerEvent(CommonSocketIOEvents.ELECTED_HOST, new EventCallback() {
//
//            @Override
//            public void onEvent(IOAcknowledge ack, Object... args) {
//                getSocketIO().setHost(true);
//                gameStart.setVisible(true);
//                gameStart.setDisabled(false);
//            }
//
//        });
//        registerEvent(CommonSocketIOEvents.GAME_START, new EventCallback() {
//
//            @Override
//            public void onEvent(IOAcknowledge ack, Object... args) {
//                switchToGame();
//            }
//
//        });
//        registerEvent(CommonSocketIOEvents.ANNOUNCEMENT, new EventCallback() {
//
//            @Override
//            public void onEvent(IOAcknowledge ack, Object... args) {
//                final String announcement = (String) args[0];
//                announcementLabel.setText(announcement);
//            }
//        });
//    }
//
//    private void buildBackground(Skin skin) {
//        // Adds a background texture to the stage
//        backgroundImage = new Image(new TextureRegion(new Texture(Gdx.files.classpath("backgrounds/swanBackground2.jpg"))));
//        backgroundImage.setBounds(0, 0, VIRTUAL_WIDTH, VIRTUAL_HEIGHT);
//        backgroundImage.setFillParent(true);
//        stage.addActor(backgroundImage);
//    }
//
//    private void buildTable(final Skin skin) {
//        table = new Table(skin);
//        table.defaults().left();
//        table.add(ipAddressLabel).padRight(LABEL_FIELD_PADDING);
//        table.add(ipAddressField).prefWidth(FIELD_WIDTH);
//        table.row();
//
//        table.add(portLabel).padRight(LABEL_FIELD_PADDING);
//        table.add(portField).prefWidth(FIELD_WIDTH);
//        table.row();
//
//        table.add(nicknameLabel).padRight(LABEL_FIELD_PADDING);
//        table.add(nicknameField).prefWidth(FIELD_WIDTH);
//        table.row();
//
//        table.add(connectButton);
//        table.add(gameStart);
//        table.row();
//        table.add(updateButton).colspan(2);
//        table.row();
//        table.add(waitingText).colspan(2);
//        table.row();
//        table.add(announcementLabel).colspan(2);
//        table.center();
//
//        table.setFillParent(true);
//    }
//
//    public void connect() {
//        final String address = SwanUtil.toAddress(ipAddressField.getText(), portField.getText());
//        try {
//            getSocketIO().connect(address, nicknameField.getText(), false, new ConnectCallback() {
//
//                @Override
//                public void onConnect(SocketIOException ex) {
//                    if (ex != null) {
//                        connectButton.setDisabled(false);
//                        connectButton.setVisible(true);
//                        new Dialog("Connection Error", skin).text("Please try again").button("OK").show(stage);
//                    } else {
//                        connectButton.setVisible(false);
//                        ipAddressField.setDisabled(true);
//                        portField.setDisabled(true);
//                        nicknameField.setDisabled(true);
//                    }
//                }
//
//                @Override
//                public void onDisconnect() {
//                    connectButton.setText("Connect");
//                    connectButton.setDisabled(false);
//                    connectButton.setVisible(true);
//                    ipAddressField.setDisabled(false);
//                    portField.setDisabled(false);
//                }
//            });
//        } catch (MalformedURLException e) {
//            final String errorMessage = "Malformed server address";
//            Gdx.app.error(CommonLogTags.SOCKET_IO, errorMessage + " " + address);
//            connectButton.setVisible(true);
//            connectButton.setDisabled(false);
//            new Dialog("Connection Error", skin).text(errorMessage).button("OK").show(stage);
//        }
//    }
//
//    @Override
//    public void render(float delta) {
//        GameUtils.clearScreen();
//        updateGameLogic(delta);
//        doRender();
//    }
//
//    @Override
//    public void resize(int width, int height) {
//        stage.getViewport().update(width, height, true);
//        SwanUtil.resizeAllFonts(stage, fontGenerator, defaultFontSize, VIRTUAL_WIDTH, VIRTUAL_HEIGHT);
//    }
//
//    @Override
//    public void show() {
//        super.show();
//        Gdx.input.setInputProcessor(stage);
//    }
//
//    @Override
//    public void hide() {
//        super.hide();
//        announcementLabel.setText("");
//    }
//
//    @Override
//    public void dispose() {
//        stage.dispose();
//        fontGenerator.dispose();
//    }

}
