package org.cpsc538B.input;

import com.badlogic.gdx.Input;
import com.badlogic.gdx.InputAdapter;
import lombok.Getter;
import org.cpsc538B.model.Direction;

/**
 * Created by newmanne on 15/03/15.
 */
public class TronInput extends InputAdapter {

    public TronInput(Direction direction) {
        this.provisionalDirection = direction;
    }

    @Getter
    private Direction provisionalDirection;

    @Override
    public boolean keyDown(int keycode) {
        switch (keycode) {
            case Input.Keys.LEFT:
                if (!provisionalDirection.equals(Direction.RIGHT)) {
                    provisionalDirection = Direction.LEFT;
                }
                break;
            case Input.Keys.RIGHT:
                if (!provisionalDirection.equals(Direction.LEFT)) {
                    provisionalDirection = Direction.RIGHT;
                }
                break;
            case Input.Keys.UP:
                if (!provisionalDirection.equals(Direction.DOWN)) {
                    provisionalDirection = Direction.UP;
                }
                break;
            case Input.Keys.DOWN:
                if (!provisionalDirection.equals(Direction.UP)) {
                    provisionalDirection = Direction.DOWN;
                }
                break;
        }
        return true;
    }

}
