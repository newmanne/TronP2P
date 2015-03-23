package org.cpsc538B;

import com.badlogic.gdx.Game;
import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.graphics.g2d.SpriteBatch;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import lombok.Getter;
import lombok.Setter;
import org.cpsc538B.go.GoSender;
import org.cpsc538B.screens.StartScreen;

import java.util.Map;

public class TronP2PGame extends Game {

    @Getter
    private SpriteBatch spritebatch;
    @Getter
    private ShapeRenderer shapeRenderer;
    @Getter
    private Assets assets;
    @Getter
    private StartScreen startScreen;
    @Getter
    private GoSender goSender;
    @Getter
    @Setter
    private Map<String, String> nicknames;

    public final static String LOG_TAG = "TRON";
    public final static String SERVER_TAG = "SERVER";
    public final static String GO_STDOUT_TAG = "GO_STDOUT";
    public final static String GO_STDERR_TAG = "GO_STDERR";

    @Override
    public void create() {
        Gdx.app.log(LOG_TAG, "Starting game!");
        spritebatch = new SpriteBatch();
        shapeRenderer = new ShapeRenderer();
        assets = new Assets();
        startScreen = new StartScreen(this);
        goSender = new GoSender();
        setScreen(startScreen);
    }

    @Override
    public void dispose() {
        assets.dispose();
        shapeRenderer.dispose();
        spritebatch.dispose();
        goSender.dispose();
    }

}
