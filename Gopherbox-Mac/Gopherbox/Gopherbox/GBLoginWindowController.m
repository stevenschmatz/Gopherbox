//
//  GBLoginWindowController.m
//  Gopherbox
//
//  Created by Steven Schmatz on 8/3/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import "GBLoginWindowController.h"
#import "GBCursorButton.h"
#import "GBAppDelegate.h"

@interface GBLoginWindowController ()

// User interface fields
@property (weak) IBOutlet NSTextField *usernameField;
@property (weak) IBOutlet NSSecureTextField *passwordField;
@property (weak) IBOutlet GBCursorButton *loginButton;
@property (weak) IBOutlet NSProgressIndicator *loginIndicator;
@property (weak) IBOutlet GBCursorButton *signupButton;
@property (weak) IBOutlet NSTextField *signupTextField;
@property (weak) IBOutlet NSTextField *passwordIncorrectField;

- (IBAction)loginInitiated:(id)sender;
- (BOOL)verifyAccountWithUsername:(NSString *)username AndPassword:(NSString *)password;

@end

@implementation GBLoginWindowController

- (id)initWithWindow:(NSWindow *)window
{
    self = [super initWithWindow:window];
    if (self) {
        self.username = @"";
        self.password = @"";
        [self.loginIndicator setHidden:YES];
        [self.passwordIncorrectField setHidden:YES];
    }
    return self;
}

- (void)windowDidLoad
{
    [super windowDidLoad];
    
    // Implement this method to handle any initialization after your window controller's window has been loaded from its nib file.
}

// Initiates the login process
- (IBAction)loginInitiated:(id)sender {
    
    // Check verification
    NSString *username = self.usernameField.stringValue;
    self.username = username;
    NSString *password = self.passwordField.stringValue;
    self.password = password;
    BOOL verified = [self verifyAccountWithUsername:username AndPassword:password];
    self.verified = verified;
    
    GBAppDelegate *appDelegate = [[NSApplication sharedApplication] delegate];
    [appDelegate loginVerified:verified withUsername:self.username andPassword:self.password];
    
    if (verified) {
        [self.loginWindow close];
    } else {
        [self.passwordIncorrectField setHidden:NO];
        NSLog(@"Login not verified. Try again.");
    }
    
    // At this point, the server should generate a random "session token" and store it in the database. Send to client as return value of login.
    // Login with these details
}

// verifyAccountWithUsername uses the login binary to validate the user.
- (BOOL)verifyAccountWithUsername:(NSString *)username AndPassword:(NSString *)password {
    
    [self.loginIndicator setHidden: NO];
    [self.loginIndicator startAnimation:self];
    
    NSString *binaryName = @"ClientValidation";
    
    NSArray *arguments = [NSArray arrayWithObjects:username, password, nil];
    
    GBAppDelegate *appDelegate = [[NSApplication sharedApplication] delegate];
    NSString *results = [appDelegate runBinaryWithName:binaryName andArguments:arguments];
    
    NSLog(@"%@", results);
    
    [self.loginIndicator stopAnimation:self];
    [self.loginIndicator setHidden: YES];
    
    
    if ([results isEqual: @"true"]) {
        return true;
    } else {
        return false;
    }
}

- (IBAction)goToSignupPage:(id)sender
{
    [[NSWorkspace sharedWorkspace] openURL: [NSURL URLWithString:@"http://gopherbox.io"]];
}

@end

