# countfloyd

A low code, local feature server platform.

What does it do?
----------------
This code provides a server communicating through a local socket accessed by a
command line tool of the same name (countfloyd). Start, populate features,
access feature instances, and stop the server when you need to. Read
configuration files to create the features you need, and then generate as many
as you want. Create your own custom constructors and/or features tailored to
your specification.


Why does this exist?
--------------------
This particular code is a partitioned offshoot of a larger tool to create
physical and video games and art. 


What is a 'constructor'?
------------------------
A constructor is a basic strategy for feature
generation. One constructor can be used to generate many and varied feature
instances. Constructors occupy a basic space: "a value corresponding to a key",
"a random color from a list", "combinatorily generated list of all items from
several other list", etc. et al. For example the simple random constructor can
used to randomize generation of features for number, color, size, etc. All
different features, generated through the same constructor type.


What is a 'feature'?
--------------------
A feature could be anything generated you need a specific instance of. This is
as simple as a value returned by a key or as complex as a detailed genetic code.


What is a 'component'?
----------------------
A component is an aggregate of features. A face features for eyes, ears, nose,
shape, etc. could be a component while an arm of features for length,
circumference, articulation, bone count, etc. could be another.


What is an 'entity'?
--------------------
An entity is an aggregate of components. Components for head, arms, torso, and
legs components could be grouped as an entity. 
