package org.cpsc538B;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.scenes.scene2d.InputEvent;
import com.badlogic.gdx.scenes.scene2d.Stage;
import com.badlogic.gdx.scenes.scene2d.ui.Label;
import com.badlogic.gdx.scenes.scene2d.ui.Table;
import com.badlogic.gdx.scenes.scene2d.ui.TextField;
import com.badlogic.gdx.scenes.scene2d.utils.ClickListener;
import com.badlogic.gdx.utils.viewport.StretchViewport;


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
        Label label = new Label("TRON", game.getAssets().getLargeLabelStyle());
        rootTable.add(label);
        label.addListener(new ClickListener() {
            @Override
            public void clicked(InputEvent event, float x, float y) {
                StartScreen.this.game.setScreen(new GameScreen(StartScreen.this.game, new PositionAndDirection(500, 500, GameScreen.Direction.DOWN), 1));
            }
        });
        rootTable.row();
        TextField ipField = new TextField("IP", game.getAssets().getSkin());
        rootTable.add(ipField);
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
