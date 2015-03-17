package org.cpsc538B.screens;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.ScreenAdapter;
import com.badlogic.gdx.graphics.Color;
import com.badlogic.gdx.graphics.glutils.ShapeRenderer;
import com.badlogic.gdx.math.Vector2;
import com.badlogic.gdx.utils.viewport.StretchViewport;
import com.google.common.collect.ImmutableMap;
import org.cpsc538B.*;
import org.cpsc538B.go.GoSender;
import org.cpsc538B.input.TronInput;
import org.cpsc538B.model.Direction;
import org.cpsc538B.model.PositionAndDirection;
import org.cpsc538B.utils.GameUtils;

import java.util.Arrays;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

/**
 * Created by newmanne on 12/03/15.
 */
public class GameScreen extends ScreenAdapter {

    // resolution
    public static final int V_WIDTH = 1920;
    public static final int V_HEIGHT = 1080;

    public static final int UNOCCUPIED = 0;

    // grid dimensions
    public final int GRID_WIDTH = 200;
    public final int GRID_HEIGHT = 200;

    // display size of grid (how big each square is)
    public final static int GRID_SIZE = 10;

    // player colors
    private final ImmutableMap<Integer, Color> pidToColor = ImmutableMap.<Integer, Color>builder()
            .put(1, Color.RED)
            .put(2, Color.BLUE)
            .put(3, Color.GREEN)
            .put(4, Color.PURPLE)
            .put(5, Color.GRAY)
            .put(6, Color.ORANGE)
            .put(7, Color.OLIVE)
            .put(8, Color.MAGENTA)
            .build();

    // libgdx stuff
    private final TronP2PGame game;
    private final StretchViewport viewport;
    private TronInput tronInput;

    // game state
    private final Map<Integer, PositionAndDirection> playerPositions;
    private final int[][] grid = new int[GRID_WIDTH][GRID_HEIGHT];
    private final int pid;

    private float accumulator;

    private final Vector2[] wallVertices = new Vector2[]{
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
        tronInput = new TronInput(startingPosition.getDirection());
        this.pid = pid;
        viewport = new StretchViewport(V_WIDTH, V_HEIGHT);
    }

    @Override
    public void show() {
        Gdx.input.setInputProcessor(tronInput);
    }

    @Override
    public void render(float delta) {
        final Collection<Object> goEvents = game.getGoSender().getGoEvents();

        for (Object event : goEvents) {	
		if (event instanceof GoSender.RoundStartEvent) {
		    Gdx.app.log("AFSKSKFJFJKSJKSFSKFJSKJF", playerPositions.get(this.pid).getX() + " " + playerPositions.get(this.pid).getY());
                final PositionAndDirection provisionalPositionAndDirection = new PositionAndDirection(getPositionAndDirection());
                switch (tronInput.getProvisionalDirection()) {
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
		Gdx.app.log("AFSKSKFJFJKSJKSFSKFJSKJF", playerPositions.get(this.pid).getX() + " " + playerPositions.get(this.pid).getY());
            } else if (event instanceof GoSender.MovesEvent) {
                // process moves
                for (GoSender.MoveEvent moveEvent : ((GoSender.MovesEvent) event).getMoves()) {
                    PositionAndDirection move = moveEvent.getPositionAndDirection();
                    grid[move.getX()][move.getY()] = moveEvent.getPid();
		    Gdx.app.log(TronP2PGame.GO_STDOUT_TAG, "MOVE " + move.getX() + " " + move.getY());
                    playerPositions.put(pid, move);
		    Gdx.app.log("AFSKSKFJFJKSJKSFSKFJSKJF", playerPositions.get(this.pid).getX() + " " + playerPositions.get(this.pid).getY());
                }
		//game.getGoSender().sendToGo(new GoSender.MoveEvent(new PositionAndDirection(getPositionAndDirection()), pid));
		game.getGoSender().sendToGo(new GoSender.NullEvent(pid));
            } else {
                throw new IllegalStateException();
            }
        }
        GameUtils.clearScreen();

        accumulator += delta;

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

    // debug
    private void printGrid() {
        for (int[] row : grid) {
            Gdx.app.log(TronP2PGame.LOG_TAG, Arrays.toString(row));
        }
    }

    private void drawGrid(ShapeRenderer shapeRenderer) {
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        for (int i = 0; i < grid.length; i++) {
            for (int j = 0; j < grid[i].length; j++) {
                int square = grid[i][j];
                if (square != UNOCCUPIED) {
                    shapeRenderer.setColor(pidToColor.get(square));
                    shapeRenderer.rect(i * GRID_SIZE, j * GRID_SIZE, GRID_SIZE, GRID_SIZE);
                }
            }
        }
        shapeRenderer.end();
    }

    private void drawWalls(ShapeRenderer shapeRenderer) {
        shapeRenderer.begin(ShapeRenderer.ShapeType.Filled);
        shapeRenderer.setColor(Color.PINK);
        for (int i = 0; i < wallVertices.length - 1; i++) {
            shapeRenderer.rectLine(wallVertices[i], wallVertices[i + 1], GRID_SIZE * 2);
        }
        shapeRenderer.end();
    }

    @Override
    public void resize(int width, int height) {
        viewport.update(width, height, true);
        // TODO: might need to resize fonts here
    }

}
