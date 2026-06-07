package php

// NOTE: Concrete CMS is intentionally skipped.
//
// In Concrete CMS, user input arrives as a method parameter — BlockController::save($args)
// and BlockController::validate($args) receive $_POST data as the $args parameter.
// This is a structural pattern (input via method parameter) that cannot be detected
// by regex-based source matching on method calls or static calls. It would require
// tracking the framework's routing/dispatch mechanism to know that $args originates
// from $_POST, which is a Priority 6 inter-procedural analysis concern beyond the
// scope of pattern-based definitions.
