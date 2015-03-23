package org.cpsc538B.screens;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.scenes.scene2d.InputEvent;
import com.badlogic.gdx.scenes.scene2d.Stage;
import com.badlogic.gdx.scenes.scene2d.Touchable;
import com.badlogic.gdx.scenes.scene2d.ui.Label;
import com.badlogic.gdx.scenes.scene2d.ui.Table;
import com.badlogic.gdx.scenes.scene2d.ui.TextButton;
import com.badlogic.gdx.scenes.scene2d.ui.TextField;
import com.badlogic.gdx.scenes.scene2d.utils.ClickListener;
import com.badlogic.gdx.utils.viewport.StretchViewport;
import com.google.common.collect.ImmutableList;
import org.apache.commons.lang3.RandomUtils;
import org.cpsc538B.go.GoSender;
import org.cpsc538B.model.Direction;
import org.cpsc538B.model.PositionAndDirection;
import org.cpsc538B.TronP2PGame;
import org.cpsc538B.utils.GameUtils;

import java.util.List;
import java.util.Map;


/**
 * Created by newmanne on 14/03/15.
 */
public class StartScreen extends ScreenAdapter {

    public static final String DEFAULT_IP = "localhost:8081";
    public static final String START_A_GAME = "START A GAME";
    public static final String JOIN_A_GAME = "JOIN A GAME";
    public static final String CREATE_A_GAME = "CREATE A GAME";
    public static final String TRON = "TRON";
    private final Stage stage;
    private final Table rootTable;
    private final TronP2PGame game;
    private boolean readyToGo = false;
    private String pid;
    private Map<String, PositionAndDirection> startingPositions;
    final List<String> sampleNames = ImmutableList.of("Blinky", "Pacman", "Robocop", "DemonSlayer", "HAL", "ChickenLittle", "HansSolo", "Yoshi", "EcologyFan", "Ghost", "GoLeafsGo", "Batman");



    public StartScreen(TronP2PGame game) {
        this.game = game;
        stage = new Stage(new StretchViewport(GameScreen.V_WIDTH, GameScreen.V_HEIGHT), game.getSpritebatch());
        rootTable = new Table();
        rootTable.setFillParent(true);
        rootTable.defaults().pad(10f);
        stage.addActor(rootTable);

        // stuff
        Label logo = new Label(TRON, game.getAssets().getLargeLabelStyle());
        final TextField leaderIpField = new TextField(DEFAULT_IP, game.getAssets().getTextFieldStyle());
        final String defaultName = sampleNames.get(RandomUtils.nextInt(0, sampleNames.size()));
        final TextField nameField = new TextField(defaultName, game.getAssets().getTextFieldStyle());

        final TextButton startAGame = new TextButton(START_A_GAME, game.getAssets().getTextButtonStyle());
        final TextButton joinAGame = new TextButton(JOIN_A_GAME, game.getAssets().getTextButtonStyle());
        final TextButton createAGame = new TextButton(CREATE_A_GAME, game.getAssets().getTextButtonStyle());
        startAGame.setTouchable(Touchable.disabled);
        startAGame.setDisabled(true);

        createAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.getGoSender().init(leaderIpField.getText(), nameField.getText(), true, (pid1, startingPositions1, nicknames) -> {
                    // need the actual switch to happpen on the thread in the render loop unfortunately
                    StartScreen.this.pid = pid1;
                    StartScreen.this.startingPositions = startingPositions1;
                    game.setNicknames(nicknames);
                    StartScreen.this.readyToGo = true;
                });
                joinAGame.setDisabled(true);
                joinAGame.setTouchable(Touchable.disabled);
                createAGame.setDisabled(true);
                createAGame.setTouchable(Touchable.disabled);
                startAGame.setDisabled(false);
                startAGame.setTouchable(Touchable.enabled);
            }
        });
        startAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.getGoSender().sendToGo("START");
            }
        });
        joinAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.getGoSender().init(leaderIpField.getText(), nameField.getText(), false, (pid1, startingPositions1, nicknames) -> {
                    // need the actual switch to happpen on the thread in the render loop unfortunately
                    StartScreen.this.pid = pid1;
                    StartScreen.this.startingPositions = startingPositions1;
                    game.setNicknames(nicknames);
                    StartScreen.this.readyToGo = true;
                });
                joinAGame.setDisabled(true);
                joinAGame.setTouchable(Touchable.disabled);
                createAGame.setDisabled(true);
                createAGame.setTouchable(Touchable.disabled);
            }
        });

        // menu positioning
        rootTable.add(logo).colspan(2);
        rootTable.row();
        rootTable.add(new Label("Leader IP:PORT", game.getAssets().getLabelStyle()));
        rootTable.add(leaderIpField).width(800);
        rootTable.row();
        rootTable.add(new Label("Nickname", game.getAssets().getLabelStyle()));
        rootTable.add(nameField).width(800);
        rootTable.row();
        rootTable.add(createAGame).colspan(2);
        rootTable.row();
        rootTable.add(joinAGame).colspan(2);
        rootTable.row();
        rootTable.add(startAGame).colspan(2);
    }

    @Override
    public void show() {
        Gdx.input.setInputProcessor(stage);
    }

    @Override
    public void resize(int width, int height) {
        GameUtils.resize(stage, width, height, game);
    }

    @Override
    public void render(float delta) {
        GameUtils.clearScreen();
        update(delta);
        stage.draw();
        if (readyToGo) {
            game.setScreen(new GameScreen(game, pid, startingPositions));
        }
    }

    protected void update(float delta) {
        stage.act(delta);
    }

    @Override
    public void dispose() {
        stage.dispose();
    }

}
