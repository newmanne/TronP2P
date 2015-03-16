package org.cpsc538B.utils;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.guava.GuavaModule;
import lombok.Getter;

import java.io.IOException;

/**
 * Created by newmanne on 14/03/15.
 */
public class JSONUtils {

    @Getter
    private static final ObjectMapper mapper;

    static {
        mapper = new ObjectMapper();
        mapper.registerModule(new GuavaModule());
    }

    public static <T> T toObject(String jsonString, Class<T> klazz) {
        try {
            return mapper.readValue(jsonString, klazz);
        } catch (IOException e) {
            throw new RuntimeException("Couldn't deserialize string " + jsonString + " into type " + klazz, e);
        }
    }

    public static String toString(Object object) {
        return toString(object, false);
    }

    public static String toString(Object object, boolean pretty) {
        try {
            final String json;
            if (pretty) {
                json = mapper.writerWithDefaultPrettyPrinter().writeValueAsString(object);
            } else {
                json = mapper.writeValueAsString(object);
            }
            return json;
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Couldn't serialize object " + object, e);
        }
    }

}
