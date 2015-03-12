package org.cpsc538B;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.Input;
import com.badlogic.gdx.InputAdapter;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.graphics.Color;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import com.badlogic.gdx.utils.viewport.StretchViewport;

/**
 * Created by newmanne on 12/03/15.
 */
public class GameScreen extends ScreenAdapter {

    private final TronP2PGame game;
    private float accumulator;
    public static final int V_WIDTH = 1920;
    public static final int V_HEIGHT = 1080;

    private final StretchViewport viewport;

    public static enum Direction {LEFT, RIGHT, DOWN, UP};
    private Direction currentDirection = Direction.DOWN;
    private int x = 500;
    private int y = 500;

    public GameScreen(TronP2PGame game) {
        this.game = game;
        viewport = new StretchViewport(V_WIDTH, V_HEIGHT);
    }

    @Override
    public void show() {
        Gdx.input.setInputProcessor(new InputAdapter() {
            @Override
            public boolean keyDown(int keycode) {
                switch (keycode) {
                    case Input.Keys.LEFT:
                        currentDirection = Direction.LEFT;
                        break;
                    case Input.Keys.RIGHT:
                        currentDirection = Direction.RIGHT;
                        break;
                    case Input.Keys.UP:
                        currentDirection = Direction.UP;
                        break;
                    case Input.Keys.DOWN:
                        currentDirection = Direction.DOWN;
                        break;
                }
                return true;
            }
        });
    }

    @Override
    public void render(float delta) {
        GameUtils.clearScreen();

        accumulator += delta;

        // game logic
        switch (currentDirection) {
            case LEFT:
                x -= 10;
                break;
            case RIGHT:
                x += 10;
                break;
            case DOWN:
                y -= 10;
                break;
            case UP:
                y += 10;
                break;
        }
        // render
        viewport.apply();
        final ShapeRenderer shapeRenderer = game.getShapeRenderer();
        shapeRenderer.setProjectionMatrix(viewport.getCamera().combined);
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
            shapeRenderer.setColor(Color.BLUE);
            shapeRenderer.rect(x, y, 10, 10);
        shapeRenderer.end();

        // TODO: scrolling
    }

    @Override
    public void resize(int width, int height) {
        viewport.update(width, height, true);
        // TODO: might need to resize fonts here
    }

}
