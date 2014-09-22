//
//  GBCursorButton.m
//  Gopherbox
//
//  Created by Steven Schmatz on 8/3/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import "GBCursorButton.h"

@implementation GBCursorButton

- (id)initWithFrame:(NSRect)frame
{
    self = [super initWithFrame:frame];
    if (self) {
        // Initialization code here.
    }
    return self;
}

- (void)drawRect:(NSRect)dirtyRect
{
    [super drawRect:dirtyRect];
    
    // Drawing code here.
}

// resetCursorRects just makes the cursor show that a link is clickable.
- (void)resetCursorRects
{
    [super resetCursorRects];
    [self addCursorRect:[self bounds] cursor:[NSCursor pointingHandCursor]];
}

@end
