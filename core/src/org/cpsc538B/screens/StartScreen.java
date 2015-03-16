package org.cpsc538B.screens;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.scenes.scene2d.InputEvent;
import com.badlogic.gdx.scenes.scene2d.Stage;
import com.badlogic.gdx.scenes.scene2d.ui.Label;
import com.badlogic.gdx.scenes.scene2d.ui.Table;
import com.badlogic.gdx.scenes.scene2d.ui.TextButton;
import com.badlogic.gdx.scenes.scene2d.ui.TextField;
import com.badlogic.gdx.scenes.scene2d.utils.ClickListener;
import com.badlogic.gdx.utils.viewport.StretchViewport;
import org.cpsc538B.model.Direction;
import org.cpsc538B.model.PositionAndDirection;
import org.cpsc538B.TronP2PGame;
import org.cpsc538B.utils.GameUtils;


/**
 * Created by newmanne on 14/03/15.
 */
public class StartScreen extends ScreenAdapter {

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
        Label logo = new Label("TRON", game.getAssets().getLargeLabelStyle());
        final TextField leaderIpField = new TextField("IP", game.getAssets().getSkin());
        final TextButton startAGame = new TextButton("START A GAME", game.getAssets().getSkin());
        final TextButton joinAGame = new TextButton("JOIN A GAME", game.getAssets().getSkin());

        startAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.getGoSender().init(leaderIpField.getText(), true);
                // TODO: can't really assign these positions just yet
                StartScreen.this.game.setScreen(new GameScreen(StartScreen.this.game, new PositionAndDirection(10, 10, Direction.DOWN), 1));
            }
        });
        joinAGame.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.getGoSender().init(leaderIpField.getText(), false);
                StartScreen.this.game.setScreen(new GameScreen(StartScreen.this.game, new PositionAndDirection(10, 10, Direction.DOWN), 1));
            }
        });

        // menu positioning
        rootTable.add(logo);
        rootTable.row();
        rootTable.add(leaderIpField);
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
