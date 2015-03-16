package org.cpsc538B.utils;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.graphics.GL20;
import com.badlogic.gdx.graphics.Texture;
import com.badlogic.gdx.graphics.g2d.BitmapFont;
import com.badlogic.gdx.graphics.g2d.freetype.FreeTypeFontGenerator;
import com.badlogic.gdx.scenes.scene2d.Actor;
import com.badlogic.gdx.scenes.scene2d.Group;
import com.badlogic.gdx.scenes.scene2d.Stage;
import com.badlogic.gdx.scenes.scene2d.ui.*;
import org.cpsc538B.TronP2PGame;

public class GameUtils {

    public static void clearScreen() {
        // clear screen
        Gdx.gl.glClearColor(0, 0, 0, 1);
        Gdx.gl.glClear(GL20.GL_COLOR_BUFFER_BIT);
    }

    public static void resize(Stage stage, int width, int height, TronP2PGame game) {
        stage.getViewport().update(width, height, true);
        game.getAssets().resizeStyles();
        GameUtils.resizeStageFonts(stage);
    }

    private static BitmapFont generateFont(FreeTypeFontGenerator fontGenerator, int size) {
        FreeTypeFontGenerator.FreeTypeFontParameter freeTypeFontParameter = new FreeTypeFontGenerator.FreeTypeFontParameter();
        freeTypeFontParameter.size = size;
        freeTypeFontParameter.magFilter = Texture.TextureFilter.Linear;
        freeTypeFontParameter.minFilter = Texture.TextureFilter.Linear;
        return fontGenerator.generateFont(freeTypeFontParameter);
    }

    public static BitmapFont generateVirtualFont(FreeTypeFontGenerator fontGenerator, int size, float virtualWidth, float virtualHeight) {
        float wScale = 1.0f * Gdx.graphics.getWidth() / virtualWidth;
        float hScale = 1.0f * Gdx.graphics.getHeight() / virtualHeight;
        if (wScale < 1) {
            wScale = 1;
        }
        if (hScale < 1) {
            hScale = 1;
        }
        BitmapFont generatedFont = generateFont(fontGenerator, (int) (size * Math.min(wScale, hScale)));
        generatedFont.setScale((float) (1.0 / wScale), (float) (1.0 / hScale));
        return generatedFont;
    }

    public static void resizeStageFonts(Stage stage) {
        for (Actor actor : stage.getActors()) {
            resizeActorsFont(actor);
        }
    }

    private static void resizeActorsFont(Actor actor) {
        if (actor instanceof TextField) {
            TextField textfield = (TextField) actor;
            textfield.setStyle(textfield.getStyle());
            // for some reason doing this makes it render properly
            textfield.setText(textfield.getText());
        } else if (actor instanceof Label) {
            Label label = (Label) actor;
            label.setStyle(label.getStyle());
        } else if (actor instanceof TextButton) {
            TextButton textButton = (TextButton) actor;
            textButton.setStyle(textButton.getStyle());
        } else if (actor instanceof com.badlogic.gdx.scenes.scene2d.ui.List<?>) {
            com.badlogic.gdx.scenes.scene2d.ui.List<?> list = (com.badlogic.gdx.scenes.scene2d.ui.List<?>) actor;
            list.setStyle(list.getStyle());
        } else if (actor instanceof Table) {
            Table table = (Table) actor;
            for (Cell<?> cell : table.getCells()) {
                if (cell.getActor() != null) {
                    resizeActorsFont(cell.getActor());
                }
            }
        } else if (actor instanceof Group) {
            Group group = (Group) actor;
            for (Actor groupActor : group.getChildren()) {
                resizeActorsFont(groupActor);
            }
        }
    }

}
