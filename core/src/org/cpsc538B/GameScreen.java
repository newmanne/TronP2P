package org.cpsc538B;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.Input;
import com.badlogic.gdx.InputAdapter;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.graphics.Color;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import com.badlogic.gdx.math.Vector2;
import com.badlogic.gdx.utils.viewport.StretchViewport;

/**
 * Created by newmanne on 12/03/15.
 */
public class GameScreen extends ScreenAdapter {

    public static final int PLAYER_WIDTH = 10;
    public static final int PLAYER_HEIGHT = 10;
    private final TronP2PGame game;
    private float accumulator;
    public static final int V_WIDTH = 1920;
    public static final int V_HEIGHT = 1080;

    private final StretchViewport viewport;

    public static enum Direction {LEFT, RIGHT, DOWN, UP};
    private PositionAndDirection positionAndDirection = new PositionAndDirection(500, 500, Direction.DOWN);

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
                        positionAndDirection.setDirection(Direction.LEFT);
                        break;
                    case Input.Keys.RIGHT:
                        positionAndDirection.setDirection(Direction.RIGHT);
                        break;
                    case Input.Keys.UP:
                        positionAndDirection.setDirection(Direction.UP);
                        break;
                    case Input.Keys.DOWN:
                        positionAndDirection.setDirection(Direction.DOWN);
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
        switch (positionAndDirection.getDirection()) {
            case LEFT:
                positionAndDirection.setX(positionAndDirection.getX() - 10);
                break;
            case RIGHT:
                positionAndDirection.setX(positionAndDirection.getX() + 10);
                break;
            case DOWN:
                positionAndDirection.setY(positionAndDirection.getY() - 10);
                break;
            case UP:
                positionAndDirection.setY(positionAndDirection.getY() + 10);
                break;
        }

        // scroll
        viewport.getCamera().position.set(positionAndDirection.getX() , positionAndDirection.getY(), 0);

        // render
        viewport.apply();
        final ShapeRenderer shapeRenderer = game.getShapeRenderer();
        shapeRenderer.setProjectionMatrix(viewport.getCamera().combined);

        drawWalls(shapeRenderer);

        // draw player
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
            shapeRenderer.setColor(Color.BLUE);
            shapeRenderer.rect(positionAndDirection.getX(), positionAndDirection.getY(), PLAYER_WIDTH, PLAYER_HEIGHT);
        shapeRenderer.end();
    }

    private void drawWalls(ShapeRenderer shapeRenderer) {
        final int WALL_DRAW_THICKNESS = 10;
        Vector2[] wallVertices = new Vector2[] {
                new Vector2((-V_WIDTH / 2), (-V_HEIGHT / 2)),
                new Vector2((V_WIDTH / 2), (-V_HEIGHT / 2)),
                new Vector2((V_WIDTH / 2), (V_HEIGHT / 2)),
                new Vector2((-V_WIDTH / 2), (V_HEIGHT / 2)),
                new Vector2((-V_WIDTH / 2), (-V_HEIGHT / 2))
        };
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        shapeRenderer.setColor(Color.WHITE);
        for (int i = 0; i < wallVertices.length - 1; i++) {
            // add a little bit extra to make sure the walls get fully drawn
            final Vector2 addition = wallVertices[i+1].cpy().sub(wallVertices[i]).nor().scl(WALL_DRAW_THICKNESS / 2);
            shapeRenderer.rectLine(wallVertices[i], wallVertices[i + 1].cpy().add(addition), WALL_DRAW_THICKNESS);
        }
        shapeRenderer.end();
    }

    @Override
    public void resize(int width, int height) {
        viewport.update(width, height, true);
        // TODO: might need to resize fonts here
    }

}
