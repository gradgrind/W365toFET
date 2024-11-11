# Internal Data Representation – Considerations

The basic data can be stored in various formats, though JSON is assumed. When it is processed, additional detail will be needed. The references to other elements should perhaps be pointers rather than keys (strings).

A particular concern is data integrity. Especially when an element is deleted there is the possibility of invalidating references.

## Tracing references

The "simplest" approach might be to search all (relevant) elements for references to a node whenever this information is needed. Although a bit clumsy, this may not be a bad choice – if deletes are rare, which is likely.

Another possibility would be to use an interface which tracks the references. This could involve "smart" pointers.

### Smart Pointers

Instead of a direct reference to the target (by key or as pointer), one could use an elaborated pointer. The simplest could be to use a reference counter, allowing a deletion only if there are no references. The disadvantage is that there is no link to the referring element. Using a list of Referrer references could overcome this shortcoming.