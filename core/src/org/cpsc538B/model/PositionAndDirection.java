package org.cpsc538B.model;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class PositionAndDirection {

    public PositionAndDirection(PositionAndDirection other) {
        this.x = other.getX();
        this.y = other.getY();
        this.direction = other.getDirection();
    }

    public PositionAndDirection(int x, int y, Direction direction) {
        this.x = x;
        this.y = y;
        this.direction = direction;
    }

    private int x;
    private int y;
    private Direction direction;
}
