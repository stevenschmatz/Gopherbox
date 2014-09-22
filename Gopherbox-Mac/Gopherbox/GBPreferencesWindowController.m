//
//  GBPreferencesWindowController.m
//  Gopherbox
//
//  Created by Steven Schmatz on 8/8/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import "GBPreferencesWindowController.h"

@interface GBPreferencesWindowController ()

@end

@implementation GBPreferencesWindowController

- (id)initWithWindow:(NSWindow *)window
{
    self = [super initWithWindow:window];
    if (self) {
    }
    return self;
}

- (void)windowDidLoad
{
    [super windowDidLoad];
    
    [[self.saveButton cell] setHighlighted:YES];
    
    NSString *folderLocationString = @"The folder is ";
    NSString *path = [[NSUserDefaults standardUserDefaults] valueForKey:@"gopherboxFolderPath"];
    NSString *finalString = [folderLocationString stringByAppendingString:path];
    
    [self.pathString setStringValue:finalString];
    // Implement this method to handle any initialization after your window controller's window has been loaded from its nib file.
}
- (IBAction)saveButtonPressed:(id)sender {
    [self.window close];
}
- (IBAction)selectFolder:(id)sender {
    // create an open documet panel
    NSOpenPanel *panel = [NSOpenPanel openPanel];
    
    panel.canChooseFiles = NO;
    panel.canChooseDirectories = YES;
    
    // display the panel
    [panel beginWithCompletionHandler:^(NSInteger result) {
        if (result == NSFileHandlingPanelOKButton) {
            
            // grab a reference to what has been selected
            NSURL *theDocument = [[panel URLs]objectAtIndex:0];
            
            // write our file name to a label
            NSString *theString = [NSString stringWithFormat:@"%@", theDocument];
            
            [[NSUserDefaults standardUserDefaults] setObject:theString forKey:@"gopherboxFolderPath"];
            [[NSUserDefaults standardUserDefaults] synchronize];

            [self.pathString setStringValue:theString];
        }
    }];
}

@end
