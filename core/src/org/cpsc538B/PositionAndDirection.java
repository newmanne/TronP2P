package org.cpsc538B;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
public class PositionAndDirection {

    PositionAndDirection(int x, int y, GameScreen.Direction direction) {
        this.x = x;
        this.y = y;
        this.direction = direction;
    }

    private int x;
    private int y;
    private GameScreen.Direction direction;
}
