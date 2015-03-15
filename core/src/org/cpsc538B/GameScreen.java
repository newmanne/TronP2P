package org.cpsc538B;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.Input;
import com.badlogic.gdx.InputAdapter;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.graphics.Color;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import com.badlogic.gdx.math.Vector2;
import com.badlogic.gdx.utils.viewport.StretchViewport;
import com.google.common.collect.ImmutableMap;

import java.util.Collection;

/**
 * Created by newmanne on 12/03/15.
 */
public class GameScreen extends ScreenAdapter {

    private static final int UNOCCUPIED = 0;
    private final TronP2PGame game;

    // resolution
    public static final int V_WIDTH = 1920;
    public static final int V_HEIGHT = 1080;

    // grid dimensions
    private final int GRID_WIDTH = 600;
    private final int GRID_HEIGHT = 600;
    private final int[][] grid = new int[GRID_WIDTH][GRID_HEIGHT];

    // display size of grid
    private final static int GRID_SIZE = 10;

    private final ImmutableMap<Integer, Color> pidToColor = ImmutableMap.of(1, Color.RED, 2, Color.BLUE);

    private final int pid;

    private final StretchViewport viewport;

    public static enum Direction {LEFT, RIGHT, DOWN, UP}

    ;
    private PositionAndDirection positionAndDirection;

    private float accumulator;
    private PositionAndDirection provisionalPositionAndDirection;

    final int WALL_DRAW_THICKNESS = 10;
    final Vector2[] wallVertices = new Vector2[]{
            new Vector2(0, 0),
            new Vector2(GRID_HEIGHT * GRID_SIZE, 0),
            new Vector2(GRID_WIDTH * GRID_SIZE, GRID_HEIGHT * GRID_SIZE),
            new Vector2(0, GRID_HEIGHT * GRID_SIZE),
            new Vector2(0, 0)
    };

    public GameScreen(TronP2PGame game, PositionAndDirection startingPosition, int pid) {
        this.game = game;
        this.positionAndDirection = startingPosition;
        provisionalPositionAndDirection = new PositionAndDirection(this.positionAndDirection);
        this.pid = pid;
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
        final Collection<Object> goEvents = game.getGoSender().getGoEvents();
        for (Object event : goEvents) {
            if (event instanceof GoSender.RoundStartEvent) {
                game.getGoSender().sendToGo(new GoSender.MoveEvent(positionAndDirection, pid));
            } else if (event instanceof GoSender.MovesEvent) {
                // process moves
                for (GoSender.MoveEvent moveEvent : ((GoSender.MovesEvent) event).getMoveEvents()) {
                    PositionAndDirection move = moveEvent.getPositionAndDirection();
                    grid[move.getX()][move.getX()] = moveEvent.getPid();
                    // if the move is me
                    if (moveEvent.getPid() == pid) {
                        positionAndDirection = moveEvent.getPositionAndDirection();
                    }
                }
            } else {
                throw new IllegalStateException();
            }
        }

        GameUtils.clearScreen();

        accumulator += delta;

        // game logic
        // figure out what the next move should be. This doesn't actually move you.
        switch (positionAndDirection.getDirection()) {
            case LEFT:
                provisionalPositionAndDirection.setX(Math.max(0, positionAndDirection.getX() - 1));
                break;
            case RIGHT:
                provisionalPositionAndDirection.setX(Math.min(GRID_WIDTH - 1, positionAndDirection.getX() + 1));
                break;
            case DOWN:
                provisionalPositionAndDirection.setY(Math.max(0, positionAndDirection.getY() - 1));
                break;
            case UP:
                provisionalPositionAndDirection.setY(Math.min(GRID_HEIGHT - 1, positionAndDirection.getY() + 1));
                break;
        }

        // scroll
        viewport.getCamera().position.set(Math.min(GRID_WIDTH * GRID_SIZE - V_WIDTH / 2, Math.max(V_WIDTH / 2, positionAndDirection.getX() * GRID_SIZE)), positionAndDirection.getY() * GRID_SIZE, 0);

        // render
        viewport.apply();
        final ShapeRenderer shapeRenderer = game.getShapeRenderer();
        shapeRenderer.setProjectionMatrix(viewport.getCamera().combined);

        drawWalls(shapeRenderer);
        drawGrid(shapeRenderer);
        // draw player
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        shapeRenderer.setColor(Color.WHITE);
        shapeRenderer.rect(positionAndDirection.getX() * GRID_SIZE, positionAndDirection.getY() * GRID_SIZE, GRID_SIZE, GRID_SIZE);
        shapeRenderer.end();
    }

    private void drawGrid(ShapeRenderer shapeRenderer) {
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        for (int i = 0; i < grid.length; i++) {
            int[] row = grid[i];
            for (int j = 0; j < row.length; j++) {
                int square = row[j];
                if (square != 0) {
                    shapeRenderer.setColor(pidToColor.get(square));
                    shapeRenderer.rect(i * GRID_SIZE, j * GRID_SIZE, GRID_SIZE, GRID_SIZE);
                }
            }
        }
        shapeRenderer.end();
    }

    private void drawWalls(ShapeRenderer shapeRenderer) {
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        shapeRenderer.setColor(Color.WHITE);
        for (int i = 0; i < wallVertices.length - 1; i++) {
            // add a little bit extra to make sure the walls get fully drawn
            final Vector2 addition = wallVertices[i + 1].cpy().sub(wallVertices[i]).nor().scl(WALL_DRAW_THICKNESS / 2);
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