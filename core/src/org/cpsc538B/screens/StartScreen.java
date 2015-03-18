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
import org.cpsc538B.go.GoSender;
import org.cpsc538B.model.Direction;
import org.cpsc538B.model.PositionAndDirection;
import org.cpsc538B.TronP2PGame;
import org.cpsc538B.utils.GameUtils;

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

    public StartScreen(TronP2PGame game) {
        this.game = game;
        stage = new Stage(new StretchViewport(GameScreen.V_WIDTH, GameScreen.V_HEIGHT), game.getSpritebatch());
        rootTable = new Table();
        rootTable.setFillParent(true);
        stage.addActor(rootTable);

        // stuff
        Label logo = new Label(TRON, game.getAssets().getLargeLabelStyle());
        final TextField leaderIpField = new TextField(DEFAULT_IP, game.getAssets().getTextFieldStyle());

        final TextButton startAGame = new TextButton(START_A_GAME, game.getAssets().getTextButtonStyle());
        final TextButton joinAGame = new TextButton(JOIN_A_GAME, game.getAssets().getTextButtonStyle());
        final TextButton createAGame = new TextButton(CREATE_A_GAME, game.getAssets().getTextButtonStyle());
        startAGame.setTouchable(Touchable.disabled);
        startAGame.setDisabled(true);

        createAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.getGoSender().init(leaderIpField.getText(), true, new GoSender.GoInitializedCallback() {
                    @Override
                    public void onGoStarted() {
                        startAGame.setDisabled(false);
                        startAGame.setTouchable(Touchable.enabled);
                    }

                    @Override
                    public void onGameStarted(int pid, Map<Integer, PositionAndDirection> startingPositions) {
                        StartScreen.this.game.setScreen(new GameScreen(StartScreen.this.game, pid, startingPositions));
                    }
                });
                joinAGame.setDisabled(true);
                joinAGame.setTouchable(Touchable.disabled);
                createAGame.setDisabled(true);
                createAGame.setTouchable(Touchable.disabled);
            }
        });
        startAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                // TODO: send a "let's actually start" event
                // StartScreen.this.game.getGoSender().sendToGo();
            }
        });
        joinAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
            }
        });

        // menu positioning
        rootTable.add(logo);
        rootTable.row();
        rootTable.add(leaderIpField).fill();
        rootTable.row();
        rootTable.add(createAGame);
        rootTable.row();
        rootTable.add(joinAGame);
        rootTable.row();
        rootTable.add(startAGame);
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
    }

    protected void update(float delta) {
        stage.act(delta);
    }

    @Override
    public void dispose() {
        stage.dispose();
    }

}
