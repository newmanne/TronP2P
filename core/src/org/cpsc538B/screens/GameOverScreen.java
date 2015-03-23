package org.cpsc538B.screens;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.scenes.scene2d.Stage;
import com.badlogic.gdx.scenes.scene2d.ui.Label;
import com.badlogic.gdx.scenes.scene2d.ui.Table;
import com.badlogic.gdx.utils.viewport.StretchViewport;
import com.google.common.collect.Lists;
import org.cpsc538B.TronP2PGame;
import org.cpsc538B.utils.GameUtils;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

/**
 * Created by newmanne on 22/03/15.
 */
public class GameOverScreen extends ScreenAdapter {

    private final TronP2PGame game;
    private final Stage stage;

    public GameOverScreen(TronP2PGame game, List<String> places) {
        this.game = game;
        stage = new Stage(new StretchViewport(GameScreen.V_WIDTH, GameScreen.V_HEIGHT), game.getSpritebatch());
        final Table rootTable = new Table();
        rootTable.setFillParent(true);
        stage.addActor(rootTable);

        final Label gameOver = new Label("GAME OVER", game.getAssets().getLargeLabelStyle());

        final Table positionsTable = new Table();
        positionsTable.defaults().padRight(50.0f);
        positionsTable.add(new Label("PLACE", game.getAssets().getLabelStyle()));
        positionsTable.add(new Label("ID", game.getAssets().getLabelStyle()));
        positionsTable.add(new Label("Nickname", game.getAssets().getLabelStyle()));
        positionsTable.row();
        final List<String> reversedPlaces = Lists.reverse(places);
        for (int i = 0; i < reversedPlaces.size(); i++) {
            final Label positionLabel = new Label(Integer.toString(i + 1), game.getAssets().getLabelStyle());
            final String pid = reversedPlaces.get(i);
            final Label idlabel = new Label(pid, game.getAssets().getLabelStyle());
            idlabel.setColor(GameScreen.pidToColor.get(pid));
            final Label nicknameLabel = new Label(game.getNicknames().get(pid), game.getAssets().getLabelStyle());
            nicknameLabel.setColor(GameScreen.pidToColor.get(pid));
            positionsTable.add(positionLabel);
            positionsTable.add(idlabel);
            positionsTable.add(nicknameLabel);
            positionsTable.row();
        }

        rootTable.add(gameOver);
        rootTable.row();
        rootTable.add(positionsTable);

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
