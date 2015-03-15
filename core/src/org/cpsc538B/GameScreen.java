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

import java.util.Arrays;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

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
    private final int GRID_WIDTH = 200;
    private final int GRID_HEIGHT = 200;
    private final int[][] grid = new int[GRID_WIDTH][GRID_HEIGHT];

    // display size of grid
    private final static int GRID_SIZE = 10;

    private final ImmutableMap<Integer, Color> pidToColor = ImmutableMap.of(1, Color.RED, 2, Color.BLUE);

    private final int pid;

    private final StretchViewport viewport;

    public static enum Direction {LEFT, RIGHT, DOWN, UP}

    private final Map<Integer, PositionAndDirection> playerPositions;
    private Direction provisionalDirection;

    private float accumulator;


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
        playerPositions = new HashMap<>();
        playerPositions.put(pid, startingPosition);
        this.provisionalDirection = startingPosition.getDirection();
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
                        provisionalDirection = Direction.LEFT;
                        break;
                    case Input.Keys.RIGHT:
                        provisionalDirection = Direction.RIGHT;
                        break;
                    case Input.Keys.UP:
                        provisionalDirection = Direction.UP;
                        break;
                    case Input.Keys.DOWN:
                        provisionalDirection = Direction.DOWN;
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
                final PositionAndDirection provisionalPositionAndDirection = new PositionAndDirection(getPositionAndDirection());
                switch (provisionalDirection) {
                    case LEFT:
                        provisionalPositionAndDirection.setDirection(Direction.LEFT);
                        provisionalPositionAndDirection.setX(Math.max(0, getPositionAndDirection().getX() - 1));
                        break;
                    case RIGHT:
                        provisionalPositionAndDirection.setDirection(Direction.RIGHT);
                        provisionalPositionAndDirection.setX(Math.min(GRID_WIDTH - 1, getPositionAndDirection().getX() + 1));
                        break;
                    case DOWN:
                        provisionalPositionAndDirection.setDirection(Direction.DOWN);
                        provisionalPositionAndDirection.setY(Math.max(0, getPositionAndDirection().getY() - 1));
                        break;
                    case UP:
                        provisionalPositionAndDirection.setDirection(Direction.UP);
                        provisionalPositionAndDirection.setY(Math.min(GRID_HEIGHT - 1, getPositionAndDirection().getY() + 1));
                        break;
                }
                game.getGoSender().sendToGo(new GoSender.MoveEvent(provisionalPositionAndDirection, pid));
            } else if (event instanceof GoSender.MovesEvent) {
                // process moves
                for (GoSender.MoveEvent moveEvent : ((GoSender.MovesEvent) event).getMoves()) {
                    PositionAndDirection move = moveEvent.getPositionAndDirection();
                    grid[move.getX()][move.getY()] = moveEvent.getPid();
                    playerPositions.put(pid, move);
                }
            } else {
                throw new IllegalStateException();
            }
        }
        GameUtils.clearScreen();

        accumulator += delta;

        // game logic
        // figure out what the next move should be. This doesn't actually move you.

        // scroll
        viewport.getCamera().position.set(Math.min(GRID_WIDTH * GRID_SIZE - V_WIDTH / 2, Math.max(V_WIDTH / 2, getPositionAndDirection().getX() * GRID_SIZE)),
                                          Math.min(GRID_HEIGHT * GRID_SIZE - V_HEIGHT / 2, Math.max(V_HEIGHT / 2, getPositionAndDirection().getY() * GRID_SIZE)),
                                          0);

        // render
        viewport.apply();
        final ShapeRenderer shapeRenderer = game.getShapeRenderer();
        shapeRenderer.setProjectionMatrix(viewport.getCamera().combined);

        drawWalls(shapeRenderer);
        drawGrid(shapeRenderer);
        drawPlayers(shapeRenderer);

    }

    private void drawPlayers(ShapeRenderer shapeRenderer) {
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        shapeRenderer.setColor(Color.WHITE);
        playerPositions.entrySet().stream().forEach(entry -> {
            final PositionAndDirection positionAndDirection = entry.getValue();
            shapeRenderer.rect(positionAndDirection.getX() * GRID_SIZE, positionAndDirection.getY() * GRID_SIZE, GRID_SIZE, GRID_SIZE);
        });
        shapeRenderer.end();
    }

    private PositionAndDirection getPositionAndDirection() {
        return playerPositions.get(pid);
    }

    private void printGrid() {
        for (int[] row : grid) {
            Gdx.app.log(TronP2PGame.LOG_TAG, Arrays.toString(row));
        }
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