function OnUpdate(user, meta) {
    checkForBot(user);

    if (user.is_bot) {
        dst[meta.id] = user;
        log("potential bot - user_id:", meta.id, " user blob:", user);
    }
}

function OnDelete(meta) {
}

function checkForBot(user) {
    // valid game score should be between 0-600
    if (user.score > 600 || user.current_level > 1000) {
        user.is_bot = true;
        return;
    }

    // no real money was spent and game coin count is too high
    if (user.money_spent === 0 && user.coins > 1000000) {
        user.is_bot = true;
        return;
    }

    // average game credits per game play session is too high
    if (user.game_credits/user.session_count > 1000) {
        user.is_bot = true;
        return;
    }

    // average coins per session too high
    if (user.coins/user.session_count > 10000) {
        user.is_bot = true;
        return;
    }
}
