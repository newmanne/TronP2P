package org.cpsc538B;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class PositionAndDirection {

    PositionAndDirection(PositionAndDirection other) {
        this.x = other.getX();
        this.y = other.getY();
        this.direction = other.getDirection();
    }

    PositionAndDirection(int x, int y, GameScreen.Direction direction) {
        this.x = x;
        this.y = y;
        this.direction = direction;
    }

    private int x;
    private int y;
    private GameScreen.Direction direction;
}
