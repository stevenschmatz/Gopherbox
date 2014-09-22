//
//  GBPreferencesWindowController.h
//  Gopherbox
//
//  Created by Steven Schmatz on 8/8/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import <Cocoa/Cocoa.h>

@interface GBPreferencesWindowController : NSWindowController

@property (weak) IBOutlet NSButton *saveButton;
@property (weak) IBOutlet NSTextField *pathString;

@end
