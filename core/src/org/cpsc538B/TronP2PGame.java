package org.cpsc538B;

import com.badlogic.gdx.ApplicationAdapter;
import com.badlogic.gdx.Game;
import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.graphics.GL20;
import com.badlogic.gdx.graphics.Texture;
import com.badlogic.gdx.graphics.g2d.SpriteBatch;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import lombok.Getter;

public class TronP2PGame extends Game {

    @Getter
    private SpriteBatch spritebatch;
    @Getter
    private ShapeRenderer shapeRenderer;
    @Getter
    private Assets assets;

    public final static String LOG_TAG = "TRON";

    @Override
    public void create() {
        Gdx.app.log(LOG_TAG, "Starting game!");
        spritebatch = new SpriteBatch();
        shapeRenderer = new ShapeRenderer();
        assets = new Assets();
        setScreen(new GameScreen(this));
    }

    @Override
    public void dispose() {
        assets.dispose();
        shapeRenderer.dispose();
        spritebatch.dispose();
    }

}
