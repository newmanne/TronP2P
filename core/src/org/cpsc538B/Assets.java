package org.cpsc538B;

import com.badlogic.gdx.Gdx;
import com.badlogic.gdx.assets.AssetManager;
import com.badlogic.gdx.graphics.Color;
import com.badlogic.gdx.graphics.g2d.BitmapFont;
import com.badlogic.gdx.graphics.g2d.freetype.FreeTypeFontGenerator;
import com.badlogic.gdx.scenes.scene2d.ui.Label.LabelStyle;
import com.badlogic.gdx.scenes.scene2d.ui.Skin;
import lombok.Data;

@Data
public class Assets {

    public final static int DEFAULT_FONT_SIZE = 32;
    public final static int LARGE_FONT_SIZE = 86;

    public static String SKIN_FILE = "skins/uiskin.json";

    BitmapFont font;
    BitmapFont largeFont;
    LabelStyle labelStyle;
    LabelStyle largeLabelStyle;
    private FreeTypeFontGenerator generator;
    private AssetManager assetManager;
    private Skin skin;

    public Assets() {
        generator = new FreeTypeFontGenerator(Gdx.files.internal("arial.ttf"));
        font = GameUtils.generateVirtualFont(generator, DEFAULT_FONT_SIZE, GameScreen.V_WIDTH, GameScreen.V_HEIGHT);
        largeFont = GameUtils.generateVirtualFont(generator, LARGE_FONT_SIZE, GameScreen.V_WIDTH, GameScreen.V_HEIGHT);
        labelStyle = new LabelStyle(font, Color.WHITE);
        largeLabelStyle = new LabelStyle(largeFont, Color.WHITE);
        assetManager = new AssetManager();

        // do any asset loading here
        assetManager.load(Gdx.files.internal(SKIN_FILE).path(), Skin.class);
        assetManager.finishLoading();
        skin = assetManager.get("skins/uiskin.json", Skin.class);
    }

    public void resizeStyles() {
        font = GameUtils.generateVirtualFont(generator, DEFAULT_FONT_SIZE, GameScreen.V_WIDTH, GameScreen.V_HEIGHT);
        labelStyle.font = font;
    }

    public void dispose() {
        font.dispose();
        largeFont.dispose();
        generator.dispose();
        assetManager.dispose();
    }

}